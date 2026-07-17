# Task 3.0: 事件分发器迁移到 JetStream

<critical>Read the `prd.md` and `tech-spec.md` files in this folder. If you do not read these files, your task will be considered invalid.</critical>

## Overview

将三个领域事件分发器从 Redis Pub/Sub 迁移到 JetStream 发布，接口 `ports.*EventDispatcher` 保持不变，前端 WS 消息载荷格式不变。这是通知链路迁移中风险最低的一环，先行验证发布通路。依赖 1.0。

<requirements>
- 用 `JetStreamAuctionStartedEventDispatcher` / `JetStreamBidPlacedEventDispatcher` / `JetStreamAuctionEndedEventDispatcher` 替换对应 `Redis*` 实现。
- 保持 `ports.AuctionStartedEventDispatcher` / `BidPlacedEventDispatcher` / `AuctionEndedEventDispatcher` 接口签名不变。
- `channel.go::BuildAuctionEventChannel` 返回点式 subject `auction.evt.{auctionID}`（发布到 `AUCTION_EVENTS` 流）。
- 复用现有 `*Payload` 结构与 JSON 序列化，保证 WS 契约不变。
- 更新 `auction/module.go` 中三个 `fx.Provide` 指向新实现。
</requirements>

## Subtasks

* [ ] 3.1 修改 `channel.go`：`auction:{id}:events` → `auction.evt.{id}`
* [ ] 3.2 重写三个 dispatcher：注入 `jetstream.JetStream`，`Publish(subject, payload)`，日志字段沿用现状
* [ ] 3.3 更新 `auction/module.go` 的三处 `fx.Annotate/As` 绑定
* [ ] 3.4 迁移/更新三个 `*_dispatcher_test.go`：改为断言 JetStream Publish（mock）而非 Redis
* [ ] 3.5 通过 mockery 为所需 JetStream 发布抽象生成 mock（如引入本地接口封装 Publish）

## Implementation Details

见 `tech-spec.md` 的 “Core Interfaces（事件分发器接口不变）”“Subject 与 Stream 设计” 与 “Testing Approach / Unit Tests（断言 subject、payload）”。为便于单测，建议在 infra 定义最小 `EventPublisher` 接口封装 `Publish`，dispatcher 依赖它而非直接依赖具体 JS 客户端。

## Success Criteria

* 三个分发器发布到 `auction.evt.{id}` 且 payload 与迁移前逐字节一致。
* 单测覆盖成功发布与序列化失败路径；`make lint test` 通过。
* 未引入对 Redis 的新依赖。

## Relevant Files

* 修改：`internal/modules/auction/infra/event/dispatcher/{channel.go,auction_started_event_dispatcher.go,bid_placed_event_dispatcher.go,auction_ended_event_dispatcher.go}`（及各自 `_test.go`）
* 修改：`internal/modules/auction/module.go`
* 参考：`internal/modules/auction/ports/event_dispatcher.go`
