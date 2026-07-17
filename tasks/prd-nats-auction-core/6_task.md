# Task 6.0: Bid Processor 事件驱动出价消费者

<critical>Read the `prd.md` and `tech-spec.md` files in this folder. If you do not read these files, your task will be considered invalid.</critical>

## Overview

实现竞拍核心：从 `AUCTION_COMMANDS` 拉取出价命令的持久化消费者，复用现有 UoW 事务逻辑落库（`FOR UPDATE` + 触发器保证正确性），落库成功后发布 `bid.placed` 领域事件；处理失败经有限重试，最终失败进入 DLQ。依赖 2.0、3.0、5.0。

<requirements>
- 新增 Bid Processor：durable pull consumer，订阅 `auction.cmd.bid.*`。
- 消费逻辑复用出价事务：`FindByID` → 域校验 `auction.PlaceBid` → 持久化 bid（含 `idempotency_key`）→ 更新 auction → `Complete`。
- 消费幂等：唯一约束 `(auction_id, idempotency_key)` 冲突视为“已处理”并 `Ack`，不重复出价。
- 落库成功后调用 3.0 的 `BidPlacedEventDispatcher` 发 `bid.placed` 到 `AUCTION_EVENTS`。
- 出价公平性：同一 auction 串行处理（按 subject 过滤消费者 + `MaxAckPending=1`）；跨 auction 并行。
- 失败处理：暂时性错误 `NakWithDelay`（指数退避）；超过 `MaxDeliver` 或永久性业务错误 → 发布到 `auction.dlq.{id}` 并 `Term`/`Ack`。
- 作为 fx 生命周期组件启动/停止（可在 `all`/新增 `bid-processor` 命令中运行）。
</requirements>

## Subtasks

* [ ] 6.1 新增 `internal/modules/auction/infra/messaging/bid_processor.go`：消费者创建、消费循环、优雅关闭
* [ ] 6.2 复用/抽取出价事务逻辑（与 `place_bid_command.go` 现有 UoW 流程共享，避免重复）
* [ ] 6.3 实现幂等（唯一冲突 → Ack）与领域事件发布（成功后 Dispatch）
* [ ] 6.4 实现重试/退避与 DLQ 发布（`streams.go` 的 `AUCTION_DLQ`）
* [ ] 6.5 fx 接线：Provider + OnStart 启动、OnStop drain；按需在 `cmd/` 暴露独立进程
* [ ] 6.6 单测：成功落库+发事件、唯一冲突→Ack、暂时错误→Nak、超限→DLQ

## Implementation Details

见 `tech-spec.md` 的 “出价命令与幂等（三层）”“出价顺序与 DLQ”“Component Overview（Bid Processor）” 与 “Testing Approach”。正确性锚定 Postgres 行锁，NATS 顺序仅服务公平性——即使放宽顺序也不破坏数据。

## Success Criteria

* 命令被可靠消费并落库；每条唯一命令恰好产生一条 bid（重复投递不重复出价）。
* 落库成功后 `bid.placed` 事件送达 WS 客户端（端到端）。
* 暂时错误自动重试、最终失败入 DLQ 且不丢失。
* 单测覆盖上述四类路径；`make lint test` 通过。

## Relevant Files

* 新增：`internal/modules/auction/infra/messaging/bid_processor.go`
* 参考/复用：`internal/modules/auction/application/command/place_bid_command.go`、`internal/modules/auction/infra/uow/*`、`internal/modules/auction/ports/{auction_uow.go,event_dispatcher.go}`
* 修改：`internal/modules/auction/module.go`；按需新增 `cmd/bid_processor.go`
