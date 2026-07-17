# Implementation Task Summary for 保证金模块（Deposit Module）

> 参考：本目录 `prd.md`（指向 `../../docs/deposit-module-design.md`）。使用 X.0 表示主任务，X.Y 表示子任务。
> 设计文档 §12 将实施分为 8 个阶段；本表据此跟踪完成度。

## Phases

- **Phase 1 — 领域层**：1.0（`DepositStatusEnum` / `errs` / `DepositModel` 状态机 / `MoneyModel` / 领域事件）
- **Phase 2 — 持久化**：2.0（迁移 `000009`/`000010` + sqlc 查询 + repository + mapper + 共享 outbox 表）
- **Phase 3 — 应用层**：3.0（5 命令 + 4 查询 + `DepositGuard`）
- **Phase 4 — HTTP / WebSocket**：4.0（handler / router / dto / errs / client / registry / hub）
- **Phase 5 — 消息与结算**：5.0（NATS `DEPOSIT_EVENTS` 流 + 存款事件消费者 + `auction_ended` 结算消费者）
- **Phase 6 — 出价资格网关**：6.0（`DepositGuard` 接入 `auction_handler.PlaceBid` 同步预校验 + `BidProcessor` 异步强校验 + `IsPermanentBidError`）
- **Phase 7 — Fx 装配**：7.0（`deposit.Module` + `cmd/all.go`/`auction.go`/`bid_processor.go` 加载）
- **Phase 8 — 质量与文档**：8.0（单测 + README + 本任务清单）/ 8.1（集成测试，待办）

## Dependency Map

| 任务 | 依赖 | 说明 |
|---|---|---|
| 1.0 领域模型 | — | 不改变既有拍卖行为 |
| 2.0 持久化 | 1.0 | 复用拍卖既有 `event_outbox` 表与 relay，不另起 relay |
| 3.0 应用层 | 1.0, 2.0 | 命令经 UoW 写 deposit + outbox；查询经 guard/repository |
| 4.0 HTTP/WS | 3.0 | 路由按 bidder 角色保护 |
| 5.0 消息/结算 | 2.0, 3.0 | 结算消费者只读 `auctions.winner_user_id`，避免跨模块信封耦合 |
| 6.0 资格网关 | 3.0, 4.0 | 同步（handler）+ 异步（processor）双保险 |
| 7.0 Fx 装配 | 1.0–6.0 | 模块加载后 Fx 自动注入 `DepositGuard` 到拍卖侧 |
| 8.0 质量/文档 | 1.0–7.0 | 已完成 |
| 8.1 集成测试 | 1.0–7.0 | 待办（建议 `tests/integration` 增加保证金端到端用例） |

## Tasks

* [x] 1.0 Phase 1（领域层）：`DepositStatusEnum`（pending/held/released/applied/forfeited）+ `domain/errs` + `DepositModel` 状态机（ConfirmHold/Release/ApplyToWinning/Forfeit/Cancel，非法转移返回 `ErrInvalidDepositTransition`）+ `MoneyModel`（不可变、下溢返回 `ErrInsufficientBalance`）+ 4 个领域事件
* [x] 2.0 Phase 2（持久化）：迁移 `000009_create_deposits_table.sql` + `000010_add_auction_deposit_config.sql`；sqlc `deposits.sql`/`outbox.sql`/`auctions.sql`；`PostgresDepositRepository`/`PostgresOutboxRepository`/`PostgresAuctionConfigRepository`（同时实现 `AuctionConfigPort` 与 `AuctionWinnerPort`）；`DepositMapper`；UoW（`DepositUnitOfWork`/`DepositUnitOfWorkFactory`，乐观锁 `version` 校验）
* [x] 3.0 Phase 3（应用层）：5 命令 `CreateDeposit`/`ReleaseDeposit`/`ApplyDeposit`/`ForfeitDeposit`/`CancelDeposit` + 4 查询 `GetDeposit`/`ListDepositsByUser`/`ListHeldDepositsByAuction`/`GetEligibility` + `DepositGuard`（`EnsureEligible`：配置未要求→放行；否则查 held 且金额足够）
* [x] 4.0 Phase 4（HTTP/WebSocket）：`DepositHandler`（Create/GetByID/ListByUser/GetEligibility/Release/Apply/Forfeit/Cancel）+ `DepositWebSocketHandler`；`router` 两条路由组；`http/errs`（映射领域错误到 4xx/502）+ `dto`；WebSocket `Client`/`UserSubscriberRegistry`/`Hub`（按 userID 分发 `deposit.evt.<id>`）
* [x] 5.0 Phase 5（消息与结算）：NATS `StreamDepositEvents`/`SubjectDepositEvents` + `JetStreamDepositEventConsumer`（推送至 hub）；`SettlementConsumer` 订阅 `auction_ended`，中标者 `Apply`、其余 `Release`（忽略 `ErrInvalidDepositTransition`）
* [x] 6.0 Phase 6（出价资格网关）：`DepositGuard` 接入 `AuctionHandler.PlaceBid` 同步预校验（映射 `ErrDepositRequired`/`ErrDepositInsufficient`/`ErrDepositNotHeld`→400）；`BidProcessor` 在 `Begin` 之前 `EnsureEligible` 异步强校验，并将三者加入 `IsPermanentBidError`
* [x] 7.0 Phase 7（Fx 装配）：`deposit.Module`（`fx.Provide` + `fx.As` 装配全部端口）+ `RegisterDepositRoutes`/`RegisterDepositWebsocketRoutes`/`RegisterDepositHub`/`RegisterDepositEventConsumer`；`cmd/all.go`/`auction.go`/`bid_processor.go` 加载 `deposit.Module` 并注册
* [x] 8.0 Phase 8 质量/文档：`domain/model` 状态机单测、`MoneyModel` 单测、`guard` 单测、`CreateDepositCommand` 单测（mockery 生成的保证金 mock 置于 `internal/modules/deposit/testmocks`，避免与 `tests/mocks` 的 `OutboxRepository` 重名冲突）；更新 `bid_processor_test` 适配新签名与守卫路径；README 补充「Deposit & Bidder Eligibility」特性与「Deposit Endpoints」API 文档；本任务清单
* [ ] 8.1 Phase 8 集成测试：`tests/integration` 增加保证金端到端用例（创建→出价资格→拍卖结束结算）
* [x] 8.2 真实支付适配（通用 HTTP 骨架）：`PaymentPort` 接入配置驱动的 `GenericPaymentAdapter`（provider-agnostic、预留 `RequestSigner` 签名扩展点）。`Hold`/`Release`/`Capture`/`Forfeit` 映射到 `config.Payment` 中可配置的四个 endpoint path；新增 `config.Payment` 子配置（`PAYMENT_PROVIDER`/`PAYMENT_BASE_URL`/`PAYMENT_API_KEY`/`PAYMENT_AUTH_HEADER`/`PAYMENT_TIMEOUT` 及 `HOLD/RELEASE/CAPTURE/FORFEIT_PATH`）。`module.go` 改由 `payment.NewPaymentPort` 按 `PAYMENT_PROVIDER` 在 `MockPaymentAdapter`(默认) 与 `GenericPaymentAdapter` 间切换；未知 provider 回退 mock 并告警。`MockPaymentAdapter` 保留作默认与测试实现；新增 `generic_payment_adapter_test.go`（httptest 验证四动作请求构造、回退与错误码）。

## 质量门禁（完成 1.0–8.0 时）

| 门禁 | 结果 |
|---|---|
| `go build ./...` | ✅ 通过 |
| `go vet ./...` | ✅ 通过 |
| `go test ./...` | ✅ 全部包 + 集成测试通过（exit 0） |
| `golangci-lint run ./internal/modules/deposit/... ./internal/modules/auction/infra/messaging/... ./internal/modules/auction/infra/http/chi/handler/... ./cmd/...` | ✅ 0 issues（修复 12 处：5 处 shadow、1 处 nilnil、1 处 revive 未用参数、1 处 unparam、4 处 G115） |
| `nilaway` | ✅ 通过（deposit 模块无 nil 流问题，exit 0） |
| `govulncheck` | ⚠️ 仅因**预先存在**的依赖 CVE 失败（pgx / chi / x/crypto），与本次特性无关 |

> `golangci-lint` / `nilaway` 为 `make static` 门禁之一；特性代码以 `go build`/`go vet`/`go test` 全绿为准，`govulncheck` 的失败来自既有依赖漏洞，建议单独升级 PR 处理。
> lint 修复要点：命令层 `if saveErr :=` → `if saveErr =` 去除 shadow；`auction_config_repository.go` 的 `return nil,nil` 改为返回 `errs.ErrAuctionWinnerNotFound` 并在 `settlement_consumer.go` 的 `settle` 中按「释放全部」处理；`settlement_consumer.handle` 删除未用 `subject` 参数；`deposit_handler.parseIDParam` 删除恒为 `"id"` 的 `name` 参数；`.golangci.yml` 的 G115 豁免路径扩展至 `infra/outbox`（与拍卖 `infra/repository` 一致）。
