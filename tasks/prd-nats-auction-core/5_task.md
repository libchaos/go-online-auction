# Task 5.0: 出价命令发布（HTTP 改为 202 Accepted）

<critical>Read the `prd.md` and `tech-spec.md` files in this folder. If you do not read these files, your task will be considered invalid.</critical>

## Overview

把出价从同步落库改为发布**出价命令**到 JetStream 命令流 `AUCTION_COMMANDS`。HTTP 接受层非阻塞返回 `202 Accepted` 与幂等键；实际落库由 6.0 的 Bid Processor 异步完成。依赖 1.0、2.0。

<requirements>
- 新增端口 `ports.BidCommandPublisher` 与 `BidCommand`（含 `IdempotencyKey`/`AuctionID`/`UserID`/`AmountInCents`/`IssuedAt`）。
- 新增 infra 实现：发布到 `auction.cmd.bid.{auctionID}`，设置 `Nats-Msg-Id = IdempotencyKey`（发布去重）。
- `PlaceBidCommand` 改为：轻量校验 → 生成/透传幂等键 → 发布命令 → 返回受理结果（不再直接 UoW 落库/发领域事件）。
- HTTP handler/DTO：支持 `Idempotency-Key` 请求头（缺省服务端生成 UUID），返回 `202` + `{ idempotency_key, status:"accepted" }`。
- 保留出价的输入合法性快速校验（金额>0 等域内快速失败），最终不变量仍由 6.0 + DB 保证。
- `BidCommand` 与后续 `bid.placed` 事件载荷需**携带 `idempotency_key`**：既用于 6.0 消费幂等，也用于 k6 端到端测试关联“受理→WS 通知”（见 `k6-test-plan.md`）。
</requirements>

## Subtasks

* [ ] 5.1 新增 `internal/modules/auction/ports/bid_command.go`（`BidCommandPublisher`、`BidCommand`、`BidCommandAck`）
* [ ] 5.2 新增 infra publisher：JetStream 发布 + `Nats-Msg-Id` 头
* [ ] 5.3 改造 `place_bid_command.go`：改为发布命令；调整返回类型/输出语义
* [ ] 5.4 更新 `auction_handler.go` 出价 handler 与 `dto/bid.go`：读 `Idempotency-Key`、返回 202
* [ ] 5.5 `auction/module.go` 注册 publisher；用 mockery 生成 `BidCommandPublisher` mock
* [ ] 5.6 更新 `place_bid_command_test.go`：断言发布命令与幂等键透传（不再断言落库）

## Implementation Details

见 `tech-spec.md` 的 “Core Interfaces（`BidCommandPublisher`）”“出价命令与幂等（第 1 层发布去重）”“API Endpoints（202 语义）”。注意：领域事件 `bid.placed` 的产生移到 6.0（落库成功后），5.0 不再直接 Dispatch。

## Success Criteria

* 出价接口返回 `202` 且响应含幂等键；命令成功发布到 `auction.cmd.bid.{id}` 并带 `Nats-Msg-Id`。
* 重复提交相同幂等键在去重窗口内只入队一次。
* 单测覆盖发布成功/失败与幂等键生成；`make lint test` 通过。

## Relevant Files

* 新增：`internal/modules/auction/ports/bid_command.go`、`internal/modules/auction/infra/messaging/publisher.go`
* 修改：`internal/modules/auction/application/command/place_bid_command.go`（及测试）
* 修改：`internal/modules/auction/infra/http/chi/handler/auction_handler.go`、`internal/modules/auction/infra/http/dto/bid.go`、`internal/modules/auction/module.go`
