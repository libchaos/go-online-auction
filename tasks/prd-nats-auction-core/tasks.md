# Implementation Task Summary for 以 NATS.io 作为竞拍核心与消息通知（移除 Redis）

> 参考：本目录 `prd.md` 与 `tech-spec.md`。使用 X.0 表示主任务，X.Y 表示子任务。

## Phases

- **Phase 1 — 基础设施**：1.0、2.0（可并行）
- **Phase 2 — 通知链路（低风险先行）**：3.0 → 4.0
- **Phase 3 — 竞拍核心（事件驱动出价）**：5.0 → 6.0
- **Phase 4 — 清理与验证**：7.0 → 8.0

## Dependency Map

| 任务 | 依赖 | 可并行于 |
|---|---|---|
| 1.0 NATS 基础封装与 fx 接入 | — | 2.0 |
| 2.0 DB 迁移：出价幂等键 | — | 1.0 |
| 3.0 事件分发器迁移到 JetStream | 1.0 | — |
| 4.0 WebSocket Hub 迁移到 JetStream 消费者 | 1.0, 3.0 | 5.0 |
| 5.0 出价命令发布（HTTP 202） | 1.0, 2.0 | 4.0 |
| 6.0 Bid Processor 消费者 | 2.0, 3.0, 5.0 | — |
| 7.0 移除 Redis | 3.0, 4.0, 5.0, 6.0 | — |
| 8.0 测试与可观测性 | 全部 | — |

## Tasks

* [x] 1.0 NATS 基础封装与 fx 接入
* [ ] 2.0 DB 迁移：出价幂等键
* [ ] 3.0 事件分发器迁移到 JetStream
* [ ] 4.0 WebSocket Hub 迁移到 JetStream 消费者
* [ ] 5.0 出价命令发布（HTTP 改为 202 Accepted）
* [ ] 6.0 Bid Processor 事件驱动出价消费者
* [ ] 7.0 移除 Redis
* [ ] 8.0 测试与可观测性
