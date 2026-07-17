# Task 8.0: 测试与可观测性

<critical>Read the `prd.md` and `tech-spec.md` files in this folder. If you do not read these files, your task will be considered invalid.</critical>

## Overview

补齐迁移后的自动化测试与可观测能力：内嵌 NATS 的端到端集成测试、关键单测收口，以及日志/指标，覆盖可靠性、幂等、顺序、DLQ 等 PRD 目标的可验证性。依赖全部前置任务。

<requirements>
- 集成测试：使用内嵌 NATS（`github.com/nats-io/nats-server/v2`）或 testcontainers 启动带 JetStream 的实例。
- 端到端用例：发布出价命令 → Bid Processor 落 Postgres → 事件流 → Hub 广播到测试 WS 客户端。
- 覆盖关键属性：至少一次投递不丢、重复命令仅一条 bid（幂等）、同一 auction 顺序/公平、失败入 DLQ。
- 可观测性：结构化日志（zerolog，沿用现有级别约定）与 Prometheus 指标。
- 遵循 `.agent/rules/unit-tests.md` 的测试规范；`make lint test` 全绿。
</requirements>

## Subtasks

* [ ] 8.1 搭建内嵌 NATS 测试基座（helper：启动/关闭、创建 Stream）
* [ ] 8.2 端到端集成测试：命令→落库→事件→WS 广播
* [ ] 8.3 幂等/重复投递测试：同一 `Nats-Msg-Id` 与唯一约束路径
* [ ] 8.4 顺序/公平与 DLQ 测试：串行处理、超限入 DLQ
* [ ] 8.5 指标埋点：`auction_bid_commands_published_total`、`auction_bid_commands_processed_total{result=ok|dup|dlq}`、`auction_events_published_total`、`auction_ws_broadcast_total`、消费者 `pending`/`redelivered`、publish 延迟直方图
* [ ] 8.6 日志收口：发布/消费成功 Debug、Nak/重投 Warn、失败/DLQ Error（含 auction_id/event_id/idempotency_key）

## Implementation Details

见 `tech-spec.md` 的 “Testing Approach（Unit / Integration）” 与 “Monitoring and Observability”。集成测试放在 `tests/`，与现有集成测试（`ENVIRONMENT=integration`）风格一致。

## Success Criteria

* 端到端与幂等/顺序/DLQ 测试稳定通过，可复现 PRD 的可靠性/幂等目标。
* 关键指标可被抓取，日志字段完整可检索。
* `make lint test`（含集成）全绿。

## Relevant Files

* 新增：`tests/`（NATS 集成测试与 helper）
* 修改：`internal/modules/auction/infra/messaging/*`、`websocket_hub.go`、`bid_processor.go`（指标/日志埋点）
* 参考：`.agent/rules/unit-tests.md`、`.mockery.yaml`、`Makefile`
