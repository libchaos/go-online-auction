# Technical Specification — 以 NATS.io (JetStream) 作为竞拍核心与消息通知，移除 Redis

## Executive Summary

当前系统仅把 Redis 用作**临时 Pub/Sub 通知总线**：三个事件分发器（`BidPlaced` / `AuctionStarted` / `AuctionEnded`）向 `auction:{id}:events` 频道 `PUBLISH` JSON，WebSocket `Hub` 通过 `PSubscribe("auction:*:events")` 消费并扇出给在线客户端。出价的事务权威仍是 Postgres（UoW + `FOR UPDATE` + DB 触发器）。

本方案用 **NATS JetStream** 完全替换 Redis，并将架构升级为**事件驱动出价处理**：

- **竞拍核心**：HTTP `PlaceBid` 不再同步落库，而是把出价**命令**发布到 JetStream 命令流 `AUCTION_COMMANDS`（`auction.cmd.bid.{auctionID}`）。一个持久化拉取消费者（Bid Processor）按序消费、复用现有 UoW 事务逻辑落库，成功后向事件流发布领域事件。命令消息携带幂等键，依赖 JetStream 去重窗口 + DB 唯一约束实现精确一次的业务效果，失败经有限重试后进入 DLQ。
- **消息通知**：领域事件发布到事件流 `AUCTION_EVENTS`（`auction.evt.{auctionID}`），WebSocket 网关节点以持久化消费者消费并广播到本地 `AuctionSubscriberRegistry`。

Postgres 行锁（`FOR UPDATE`）继续作为最终一致性的兜底，因此 NATS 严格顺序仅用于**出价公平性**而非正确性。整体保持现有 fx 模块化、DDD 分层与端口/适配器结构不变，改动集中在 `pkg/redis`→`pkg/nats`、shared 模块、事件分发器与 Hub。

## System Architecture

### Component Overview

```
                    ┌──────────────────────── NATS JetStream ─────────────────────────┐
HTTP PlaceBid ─────▶│  Stream AUCTION_COMMANDS  (subj: auction.cmd.bid.*, WorkQueue)   │
  (202 Accepted)    │        │                                                          │
                    │        ▼  pull (durable, MaxAckPending 调优, ordered/auction)     │
                    │   Bid Processor ── UoW(tx, FOR UPDATE) ──▶ Postgres               │
                    │        │  成功后 publish                                          │
                    │        ▼                                                          │
                    │  Stream AUCTION_EVENTS   (subj: auction.evt.*, Limits/file)       │
                    │        │                          ▲ AuctionStarted/Ended 亦发布   │
                    └────────┼──────────────────────────┼──────────────────────────────┘
                             ▼ pull (durable, per WS node)
                     WebSocket Hub ──▶ AuctionSubscriberRegistry ──▶ WS Clients
                             │
                     失败 N 次 ─▶ DLQ (auction.dlq.>)
```

主要组件与职责：

- **`pkg/nats`**：NATS 连接与 JetStream 上下文的薄封装（替代 `pkg/redis`），负责连接、TLS、重连、健康检查。
- **`internal/shared/modules/nats`**：fx Provider，暴露 `*nats.Conn` / `jetstream.JetStream`，并在生命周期钩子中**声明式创建/更新 Stream**（幂等 `CreateOrUpdateStream`）与优雅关闭。
- **Bid Command Publisher**（新端口 `ports.BidCommandPublisher`）：`PlaceBidCommand` 依赖它把出价命令发布到 `AUCTION_COMMANDS`。
- **Bid Processor**（新组件，独立可运行的 consumer）：消费命令、执行事务落库、发布领域事件、处理重试/DLQ。
- **JetStream 事件分发器**：三个分发器由 `Redis*EventDispatcher` 改为 `JetStream*EventDispatcher`，接口 `ports.*EventDispatcher` 保持不变。
- **WebSocket `Hub`**：由 `PSubscribe` 改为 JetStream 持久化消费者，其余（`register`/`unregister`/`registry.Broadcast`）不变。

## Implementation Design

### Core Interfaces

```go
// pkg/nats — 连接与 JetStream 封装
package nats

type Config struct {
    URL           string        // nats://host:4222
    Name          string        // 连接名，便于监控
    Creds         string        // 可选：NATS 凭证文件路径
    MaxReconnects int
    DedupeWindow  time.Duration // 命令流去重窗口
}

func New(cfg Config) (*natsgo.Conn, jetstream.JetStream, error)
```

```go
// internal/modules/auction/ports — 新增出价命令端口（替换同步落库入口）
type BidCommandPublisher interface {
    // Publish 发布出价命令；idempotencyKey 作为 Nats-Msg-Id 参与去重
    Publish(ctx context.Context, cmd BidCommand) (BidCommandAck, error)
}

// 事件分发器接口保持不变（仅实现从 Redis 换成 JetStream）
type BidPlacedEventDispatcher interface {
    Dispatch(ctx context.Context, event event.BidPlacedEvent) error
}
```

```go
// internal/modules/auction/infra/messaging — 事件流订阅（Hub 内部使用）
type EventConsumer interface {
    // Consume 阻塞消费 auction.evt.*，对每条消息回调 handler
    Consume(ctx context.Context, handler func(subject string, data []byte)) error
}
```

### Subject 与 Stream 设计

NATS 使用 `.` 分隔、`*`/`>` 通配。频道命名从 Redis 冒号式迁移为 NATS 令牌式：

| 用途 | Subject | Stream | 保留策略 | 存储 |
|---|---|---|---|---|
| 出价命令 | `auction.cmd.bid.{auctionID}` | `AUCTION_COMMANDS` | WorkQueue（消费即删） | file |
| 领域事件 | `auction.evt.{auctionID}` | `AUCTION_EVENTS` | Limits（TTL/MaxBytes） | file |
| 死信 | `auction.dlq.{auctionID}` | `AUCTION_DLQ` | Limits | file |

- `channel.go::BuildAuctionEventChannel` → 返回 `auction.evt.{id}`；Hub 过滤 `auction.evt.*`。
- `websocket_hub.go::extractAuctionID` 由按 `:` 切分改为按 `.` 切分（token[2]）。

### 出价命令与幂等

```go
type BidCommand struct {
    IdempotencyKey string `json:"idempotency_key"` // 客户端生成(UUID)，= Nats-Msg-Id
    AuctionID      uint64 `json:"auction_id"`
    UserID         uint64 `json:"user_id"`
    AmountInCents  uint64 `json:"amount_in_cents"`
    IssuedAt       time.Time `json:"issued_at"`
}
```

幂等三层：
1. **发布去重**：`Nats-Msg-Id = IdempotencyKey`，JetStream `DedupeWindow`（默认 2m）丢弃重复发布。
2. **消费幂等**：`bid` 表新增列 `idempotency_key`（唯一约束 `(auction_id, idempotency_key)`）；重复命令落库触发唯一冲突 → 视为已处理并 `Ack`。
3. **事务正确性**：`FOR UPDATE` + DB 触发器保证并发出价的金额单调与竞态兜底（沿用现状）。

重试与 DLQ：处理暂时性错误（DB 超时等）→ `NakWithDelay`（指数退避）；达到 `MaxDeliver` 或业务永久性错误（如无效出价）→ 发布到 `auction.dlq.{id}` 并 `Term`/`Ack`。

出价顺序：Bid Processor 采用按 `auctionID` 过滤的消费者 + `MaxAckPending=1` 保证同一场竞拍串行处理以获得公平性；跨竞拍并行。正确性由 Postgres 行锁保证，即使放宽顺序也不会破坏数据。

### Data Models

- **迁移** `migrations/`：为 `bid` 增加 `idempotency_key VARCHAR NOT NULL` 与唯一索引 `ux_bid_auction_idempotency (auction_id, idempotency_key)`。
- **消息载荷**：复用现有 `BidPlacedPayload` / `AuctionStartedPayload` / `AuctionEndedPayload`（仅发布通道变更），保证前端 WS 契约不变。

### API Endpoints

- `POST /api/v0/auctions/{id}/bids`：语义由同步 `201 Created` 改为**异步 `202 Accepted`**，返回 `{ idempotency_key, status: "accepted" }`。最终结果通过既有 WebSocket `bid.placed` 事件回推；请求头支持 `Idempotency-Key`（缺省则服务端生成）。
- WebSocket 端点（`/ws/...`）行为与消息格式不变。

## Integration Points

- **NATS 服务**：`docker-compose.yaml` 移除 `redis` 服务，新增 `nats:2.10-alpine`，命令 `-js -sd /data`（启用 JetStream + 文件存储卷）。
- **认证/TLS**：支持 `NATS_CREDS`（NGS/自建账号）与 TLS（`Password!=""` 场景等价迁移到 creds/TLS）。
- **错误处理**：连接不可用时 Publisher 返回错误 → HTTP `503`；消费者断连由 JetStream 客户端自动重连 + 未 Ack 消息重投保证不丢。

## Testing Approach

### Unit Tests

- **事件分发器**：mock `jetstream.JetStream.Publish`，断言 subject、`Nats-Msg-Id`、payload 序列化（对齐现有 `*_dispatcher_test.go`）。
- **Bid Processor**：mock UoW 与事件分发器，覆盖：成功落库并发事件、唯一冲突→Ack、暂时错误→Nak、超过 MaxDeliver→DLQ。
- **Publisher**：断言命令序列化与幂等键透传。
- 通过 `mockery`（`.mockery.yaml`，`all: true`）为新端口生成 mock；删除 `tests/mocks/mock_UniversalClient.go`。

### Integration Tests

- 使用**内嵌 NATS server**（`github.com/nats-io/nats-server/v2` 或 testcontainers）启动带 JetStream 的实例。
- 端到端：发布出价命令 → Processor 落 Postgres → 事件流 → Hub 广播到测试 WS 客户端；验证顺序、幂等（重复发布仅一条 bid）、DLQ 路径。

## Development Sequencing

### Build Order

1. **`pkg/nats` + `internal/shared/modules/nats` + config**：连接、JetStream、Stream 声明式创建。先行，是一切依赖。
2. **DB 迁移**：`bid.idempotency_key` 与唯一索引（消费者幂等前置）。
3. **事件分发器迁移**：三个 `JetStream*EventDispatcher`，接口不变，最小风险先验证发布链路。
4. **WebSocket Hub 迁移**：`PSubscribe`→JetStream 消费者，打通通知链路（此时已可端到端演示）。
5. **事件驱动出价**：`BidCommandPublisher` + Bid Processor（消费/幂等/重试/DLQ），`PlaceBidCommand` 与 HTTP 改为异步 202。
6. **移除 Redis**：删 `pkg/redis`、`internal/shared/modules/redis`、config `Redis`、`go.mod` 依赖、cmd 中 `redis.Module`、docker-compose、`.env*` 中 `REDIS_*`、mock。
7. **测试补全**：单测 + 内嵌 NATS 集成测试；`make lint test`。

### Technical Dependencies

- 依赖新增 `github.com/nats-io/nats.go`（含 `jetstream` 包）；测试可选 `github.com/nats-io/nats-server/v2`。
- 运行需可达的启用 JetStream 的 NATS（本地由 docker-compose 提供）。

## Monitoring and Observability

- **日志**（zerolog，沿用现有级别）：`Debug` 发布/消费成功（subject、auction_id、event_id、idempotency_key）；`Warn` Nak/重投；`Error` 发布失败/DLQ。
- **指标**（Prometheus 命名建议）：`auction_bid_commands_published_total`、`auction_bid_commands_processed_total{result=ok|dup|dlq}`、`auction_events_published_total`、`auction_ws_broadcast_total`、消费者 `pending`/`redelivered`、publish 延迟直方图。
- JetStream 自身可经 `nats-server` 监控端点 + Grafana NATS 面板观测 Stream/Consumer 堆积。

## Architecture Rationale & Trade-offs

本节针对**当前代码库的具体短板**说明新架构的优势，以及为获得这些优势所付出的成本，作为方案的决策依据。

### 相对现状的优势

1. **消除事件丢失与提交后双写不一致（最核心收益）**
   - 现状 `Hub.Run()` 用 Redis `PSubscribe`，Pub/Sub 即发即弃：发布时无在线订阅者（WS 节点重启、滚动部署、订阅者断连）即**永久丢事件**，看板漏更新。JetStream 持久化 + 消费者游标 + 重投，恢复后可续消费。
   - 现状 `place_bid_command.go` 先提交 UoW 再 `Dispatch()`；若 Dispatch 失败则 `return err`——**出价已落库、客户端却收到错误且事件丢失**的不一致窗口。新架构把“落库→发事件”收敛进 Bid Processor，失败走 Nak 重投而非丢弃，事件可幂等重放。

2. **接受层与数据库解耦、可削峰**
   - 现状出价同步阻塞至 `FindByID→PlaceBid→持久化→Update→Complete` 全部完成；DB 抖动直接抬高接受层延迟与失败率。
   - 新架构 HTTP 仅轻校验 + 发命令 → `202`（与 DB 无关），落库由命令流后的 Processor 按自身节奏消费；峰值排队削峰而非拒绝，接受吞吐与落库吞吐可各自扩展。

3. **顺序与公平性显式化**
   - 现状并发出价为无序竞争，正确性靠 `FOR UPDATE`（保留）。新架构“按 auction 过滤消费者 + `MaxAckPending=1`”实现同场竞拍串行按序、跨场并行；正确性仍锚定 DB 行锁，放宽顺序也不出错，风险可控。

4. **可审计 / 可回溯**：`AUCTION_EVENTS` 为可保留、可重放的持久事件序列，支撑争议出价回放、看板状态重建与合规留存——Pub/Sub 架构无法提供。

5. **失败可见（DLQ 取代静默丢失）**：现状处理失败仅一次 HTTP 500 无留痕；新架构暂时性错误自动退避重投、永久失败入 `AUCTION_DLQ`，失败可观测、可补偿。

6. **技术栈简化**：一个 NATS 同时承担命令流与通知扇出，中间件由「Redis + Postgres」变「NATS + Postgres」。

### 对应成本（非缺陷，是本架构的“账单”）

| 收益 | 成本 |
|---|---|
| 削峰/解耦 | API 变异步（202），前端需改为“受理 + 监听 WS 结果”并处理中间态/超时 |
| 至少一次投递 | 必须做幂等（`Nats-Msg-Id` + DB 唯一约束），否则重复投递=重复出价 |
| 顺序公平 | `MaxAckPending=1` 限制单场竞拍并发落库吞吐，超高频竞拍需评估分区策略 |
| 可靠/可审计 | JetStream 运维学习曲线（Stream/Consumer/DLQ/保留策略），重于 Redis Pub/Sub |

### 结论

对本代码库最实打实的两点：**(a)** 消除“事件即发即弃 + 提交后发事件失败”的双重丢失/不一致窗口；**(b)** 出价由同步阻塞改为可削峰的异步管道。可审计与公平性为顺带的结构性红利。前提是接受**异步 API + 强制幂等**这两项复杂度（已在 PRD 阶段确认接受异步），故路径自洽。

## Technical Considerations

### Key Decisions

- **JetStream 而非 Core NATS**：需要至少一次投递、按竞拍有序与可重放，以支撑“竞拍核心”的可靠性与审计；代价是需运维 Stream/Consumer。
- **事件驱动出价（异步 202）**：出价吞吐与削峰更好、天然重试；代价是客户端需经 WS 获取最终结果，API 语义从同步改为异步。
- **正确性锚定 Postgres**：保留 `FOR UPDATE` + 触发器，NATS 顺序只服务公平性，降低对严格全局有序的耦合与风险。
- **接口不变式**：`ports.*EventDispatcher` 签名不变，替换实现即可，最大化复用现有命令/测试。

### Known Risks

- **异步语义变更**影响前端/调用方 → 需 WS 回推与 `Idempotency-Key` 约定；提供过渡期文档。
- **重复投递**：JetStream 至少一次 → 必须依赖幂等键 + DB 唯一约束，否则重复出价。
- **顺序与公平性**：`MaxAckPending=1` 限制单竞拍并发吞吐 → 对超高频竞拍需评估分区策略。
- **DLQ 处理**：需人工/补偿流程消费 DLQ；纳入运维手册。
- **消息与 DB 双写一致性**：先提交事务后发事件，发事件失败会漏通知 → 采用“事务成功即 Ack 前发布，发布失败则 Nak 重投”保证最终一致（事件可幂等重放）。

### Standards Compliance

对齐 `.agent/rules`：
- `code-style-guide.md`：端口/适配器命名与 fx Provider 风格保持一致。
- `domain-model.md` / `value-object.md`：`Money` 等值对象与领域事件不变，消息层不泄漏领域不变量。
- `unit-tests.md`：新组件补齐表驱动单测与 mockery mock。
- `enum.md`：`auction_state_enum` 不受影响。

### Relevant Files

新增：
- `pkg/nats/nats.go`, `pkg/nats/config.go`
- `internal/shared/modules/nats/{nats.go,module.go}`
- `internal/shared/modules/config/nats.go`
- `internal/modules/auction/ports/bid_command.go`（`BidCommandPublisher`, `BidCommand`）
- `internal/modules/auction/infra/messaging/{publisher.go,bid_processor.go,event_consumer.go,streams.go}`
- `migrations/{n}_add_bid_idempotency_key.up/down.sql`

修改：
- `internal/modules/auction/infra/event/dispatcher/*` → JetStream 实现 + `channel.go` subject 格式
- `internal/modules/auction/infra/websocket/websocket_hub.go` → JetStream 消费者，`extractAuctionID` 按 `.` 切分
- `internal/modules/auction/module.go` → Provider 由 Redis 换 NATS，注册 Bid Processor
- `internal/modules/auction/application/command/place_bid_command.go` → 改为发布命令；HTTP handler/DTO 返回 202
- `cmd/{auction,websocket,all}.go` → `redis.Module` → `nats.Module`
- `internal/shared/modules/config/{config.go}` → `Redis` 字段替换为 `NATS`
- `docker-compose.yaml`, `.env.example`, `go.mod/go.sum`

删除：
- `pkg/redis/*`, `internal/shared/modules/redis/*`, `internal/shared/modules/config/redis.go`
- `tests/mocks/mock_UniversalClient.go`
- `.env*` 中 `REDIS_*`
