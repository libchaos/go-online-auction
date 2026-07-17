# Task 2.0: DB 迁移 — 出价幂等键

<critical>Read the `prd.md` and `tech-spec.md` files in this folder. If you do not read these files, your task will be considered invalid.</critical>

## Overview

为出价表增加幂等键与唯一约束，作为事件驱动出价处理（6.0）的消费幂等基础，确保至少一次投递下重复命令不会产生重复出价。可与 1.0 并行。

<requirements>
- 为 `bid` 表新增列 `idempotency_key VARCHAR NOT NULL`。
- 新增唯一索引 `ux_bid_auction_idempotency (auction_id, idempotency_key)`。
- 提供成对的 up/down 迁移脚本，遵循 `migrations/` 现有命名与 golang-migrate 规范。
- 更新 `bid` 相关 entity/mapper 以承载 `idempotency_key`（读写往返）。
</requirements>

## Subtasks

* [ ] 2.1 新增迁移 `migrations/{n}_add_bid_idempotency_key.up.sql` / `.down.sql`
* [ ] 2.2 在 `bid_entity.go` 增加 `IdempotencyKey` 字段；更新 `bid_mapper.go` 映射
* [ ] 2.3 评估 `bid_model.go`：幂等键属基础设施关注点，优先保留在 entity/命令层，避免污染领域模型（见 `.agent/rules/domain-model.md`）
* [ ] 2.4 本地执行迁移验证 up/down 可逆；更新受影响的 mapper 单测

## Implementation Details

见 `tech-spec.md` 的 “Data Models” 与 “出价命令与幂等”（三层幂等中的第 2 层：DB 唯一约束）。唯一冲突在 6.0 中被消费者视为“已处理”并 Ack。

## Success Criteria

* 迁移可正确 up/down；唯一索引生效，重复 `(auction_id, idempotency_key)` 插入被拒绝。
* `bid` 往返读写包含 `idempotency_key`。
* 现有仓储/映射单测通过，`make lint test` 通过。

## Relevant Files

* 新增：`migrations/{n}_add_bid_idempotency_key.up.sql`、`.down.sql`
* 修改：`internal/modules/auction/infra/entity/bid_entity.go`、`internal/modules/auction/infra/mapper/bid_mapper.go`（及其测试）
* 参考：`internal/modules/auction/infra/repository/bid_repository.go`
