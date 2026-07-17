# Task 4.0: WebSocket Hub 迁移到 JetStream 消费者

<critical>Read the `prd.md` and `tech-spec.md` files in this folder. If you do not read these files, your task will be considered invalid.</critical>

## Overview

将 WebSocket `Hub` 的事件来源从 Redis `PSubscribe("auction:*:events")` 迁移到 JetStream 持久化消费者，消费 `auction.evt.*` 并广播到本地 `AuctionSubscriberRegistry`。`register`/`unregister`/`Broadcast` 等本地扇出逻辑保持不变。依赖 1.0、3.0。

<requirements>
- `Hub` 注入 `jetstream.JetStream`（或 3.0 的 `EventConsumer` 抽象）替换 `redis.UniversalClient`/`redis.PubSub`。
- 以持久化/每 WS 节点消费者订阅 `auction.evt.*`，回调中提取 auctionID 并 `registry.Broadcast`。
- `extractAuctionID` 由按 `:` 切分改为按 `.` 切分（取 `auction.evt.{id}` 的 id 令牌）。
- `Run(ctx)` 的 select 循环与优雅关闭（`Shutdown`）语义保持等价；ctx 取消时干净退出、取消订阅。
- 更新 `NewHub` 构造与 `auction/module.go`、`RegisterWebsocketRoutes` 生命周期接线。
</requirements>

## Subtasks

* [ ] 4.1 改造 `websocket_hub.go`：字段/构造从 redis 换为 JetStream 消费者
* [ ] 4.2 用 JetStream 消费替换 `PSubscribe`；消息回调 → `Broadcast`
* [ ] 4.3 修正 `extractAuctionID`：`strings.Split(subject, ".")` 取 id 令牌 + 边界处理
* [ ] 4.4 `Shutdown`/OnStop：停止消费者、drain、无泄漏
* [ ] 4.5 更新 `auction/module.go`（Hub Provider）与相关测试

## Implementation Details

见 `tech-spec.md` 的 “Component Overview（WebSocket Hub）”“Core Interfaces（`EventConsumer`）” 与 “Subject 与 Stream 设计”。多 WS 节点场景下每节点使用独立消费者以各自获得全量事件用于本地扇出（保持现状“每节点持有本地订阅者”的语义）。

## Success Criteria

* 出价/开始/结束事件经 JetStream 可靠送达并广播到订阅该竞拍的 WS 客户端。
* 订阅者短暂离线/节点重启后关键事件不丢（相较 Redis Pub/Sub 的改进，可在 8.0 集成测试验证）。
* ctx 取消后 Hub 与消费者干净退出；`make lint test` 通过。

## Relevant Files

* 修改：`internal/modules/auction/infra/websocket/websocket_hub.go`
* 修改：`internal/modules/auction/module.go`（`NewHub` / `RegisterWebsocketRoutes`）
* 参考：`internal/modules/auction/infra/websocket/auction_subscriber_registry.go`
