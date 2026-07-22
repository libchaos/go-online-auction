# 支付服务设计文档（对接支付宝）

> 状态：设计稿（待评审，未实施）
> 日期：2026-07-20
> 参考：grayfalcon666/hunter-pay `payment-service`（仅支付宝：当面付 / 转账到支付宝账户）

## 1. 背景与目标

在现有拍卖系统中引入独立的**支付服务**，负责两类资金动作：

- **充值（用户 → 平台）**：用户通过支付宝「当面付」扫码付款，支付宝异步回调通知后，平台账本给用户账户加余额。
- **提现（平台 → 用户支付宝）**：用户申请提现，平台先冻结其账本余额，再经 MQ 异步调支付宝「转账到支付宝账户」打款；打款失败则补偿解冻。

目标：复用既有 `ledger` 账本（最终资金仲裁者）、NATS JetStream、以及 transactional outbox + relay 模式，保持六边形 + CQRS + DDD + Uber Fx 风格，与 `deposit`/`ledger` 模块平级。

## 2. 参考 hunter-pay 的关键点

- 支付服务仅集成**支付宝**：充值用当面付（`alipay.trade.precreate` 取二维码），提现用转账到支付宝账户（`alipay.fund.trans.uni.transfer`）。
- 充值：`CreatePayment` → 建单(PROCESSING) → 调支付宝取二维码 → 用户扫码 → 支付宝异步回调 `/webhook/alipay` → 验签 → 发布 `PaymentSuccessMessage` 到 RabbitMQ `payment_queue` → SimpleBank 消费 → `Transfer(平台账户 → 用户账户)` 入账。
- 提现：`CreateWithdrawal` → 建单(PROCESSING) → 发布 `WithdrawalMessage` 到 `withdrawal_queue` → **PaymentService 自己消费** → 调支付宝打款 → 成功 `WithdrawFromFrozen` 永久扣款 / 失败 `Unfreeze` 补偿解冻。
- 全链路用 `idempotency_key`（如 `out_trade_no` / `out_biz_no`）保证分布式幂等。

## 3. 本项目的适配决策（推荐）

| 维度 | 决策 | 理由 |
|------|------|------|
| 模块边界 | 新建顶层 `internal/modules/payment` | 充值/提现是独立限界上下文，不应塞进 `deposit`（押金是另一上下文）；与 `deposit`/`ledger` 平级 |
| 消息队列 | 复用 **NATS JetStream**（非 RabbitMQ） | 项目统一基础设施；采用既有 transactional outbox + relay（`payments_outbox` 表 → relay → NATS → consumer） |
| 账本 | 复用 `ledger/ports.LedgerRepository`（`Freeze/Unfreeze/WithdrawFromFrozen/Transfer/GetOrCreateAccountByOwner`） | payment 模块经依赖注入使用，不重复造轮子 |
| 支付宝对接 | 新增 `AlipayPort` 端口 + `AlipaySDKAdapter`（smartwalle/alipay/v3）+ `MockAlipayAdapter`（默认，无凭证可编译/单测/本地跑） | 端口隔离，provider 可切换；mock 保证门禁通过 |
| 现有 `deposit/infra/payment` | **保留不动** | 它是押金场景的可选外部资金动作抽象（Hold/Release/Capture/Forfeit），与充值/提现语义不同，不混用 |

## 4. 模块目录结构

```
internal/modules/payment/
  domain/
    model/      payment_model.go, withdrawal_model.go   # 聚合 + 状态机
    enum/       payment_status.go, withdrawal_status.go
    errs/       errs.go
    event/      payment_event.go                        # PaymentSuccess / WithdrawalRequested / Completed / Failed
  ports/
    payment_repository.go     # Save / FindByOutTradeNo / FindByOutBizNo / UpdateStatus
    payment_uow.go            # Begin：同一 pgx.Tx 构造 payment repo + ledger repo
    alipay_port.go            # CreateFaceToFacePayment / VerifyNotify / TransferToAlipayAccount
  application/
    command/   create_deposit_command.go, create_withdrawal_command.go, alipay_notify_command.go
    query/     get_deposit_query.go, get_withdrawal_query.go
    service/   deposit_success_consumer.go, withdrawal_consumer.go   # NATS 订阅端（Saga 执行器）
  infra/
    http/chi/{handler,router}/   payment_handler.go, payment_router.go, alipay_notify_handler.go
    http/dto/  dto.go
    repository/  pg/ postgres_payment_repository.go, postgres_withdrawal_repository.go
    uow/        wow.go, factory.go
    alipay/     sdk_adapter.go, mock_adapter.go, factory.go
    event/      envelope/  from_domain.go            # 领域事件 → outbox 信封
    outbox/     postgres_outbox_repository.go, relay.go
  module.go
```

## 5. 领域模型

**Payment 聚合**（充值单）
- 字段：`ID, UserID, AmountInCents, Currency, Status, OutTradeNo(幂等, UNIQUE), QRCodeURL, AlipayTradeNo, Version`
- 状态机：`CREATED → SUCCESS | FAILED`；方法 `MarkSuccess(tradeNo)` / `MarkFailed(reason)`（不变量校验）

**Withdrawal 聚合**（提现单）
- 字段：`ID, UserID, LedgerAccountID, AlipayAccount, AlipayRealName, AmountInCents, Currency, Status, OutBizNo(幂等, UNIQUE), FrozenOpID, AlipayOrderID, FailReason, Version`
- 状态机：`CREATED → FROZEN → SUCCESS | FAILED`；方法 `MarkFrozen(opID)` / `MarkSuccess(orderID)` / `MarkFailed(reason)`

## 6. 端口（六边形边界）

- `AlipayPort`：
  - `CreateFaceToFacePayment(ctx, FaceToFaceInput) (qrCodeURL, outTradeNo string, err error)`
  - `VerifyNotify(ctx, params map[string]string) (tradeNo, outTradeNo, tradeStatus string, err error)`
  - `TransferToAlipayAccount(ctx, TransferInput) (alipayOrderID string, err error)`
- `PaymentRepository`：`Save` / `FindByOutTradeNo` / `FindByOutBizNo` / `UpdateStatus`
- `PaymentUnitOfWork`：`Begin(ctx)` 返回在**同一 `pgx.Tx`** 上构造的 `PaymentRepository` 与 `LedgerRepository`（沿用 `deposit` 的 UoW 写法，保证支付单写与账本动作原子）
- ledger 直接注入 `ledger/ports.LedgerRepository`，不再额外抽端口

## 7. 充值流程（当面付 + 回调 + MQ + 入账）

```
前端                    payment 模块                        支付宝              NATS/ledger
  │                        │                                  │                  │
  │ POST /deposit          │                                  │                  │
  ├───────────────────────►│ CreateDepositCommand             │                  │
  │                        │  AlipayPort.CreateFaceToFacePayment               │
  │                        ├─────────────────────────────────►│ (取二维码)       │
  │                        │  保存 Payment(CREATED)            │                  │
  │◄────── qrCodeURL ──────┤                                  │                  │
  │   用户扫码支付          │                                  │                  │
  │                        │       支付宝异步回调 /alipay/notify                  │
  │                        │◄─────────────────────────────────┤                  │
  │                        │ AlipayNotifyCommand.VerifyNotify                  │
  │                        │  UoW: Payment→SUCCESS            │                  │
  │                        │  UoW: 写 payments_outbox(PaymentSuccess)          │
  │                        │            (relay 轮询) ──────────────────────────►│ payment.evt.deposit.success
  │                        │                                   │  DepositSuccessConsumer
  │                        │                                   │  UoW: ledger.Transfer(平台→用户, 幂等=out_trade_no)
```

要点：
- 回调 handler **不**直接动账本，而是更新支付单状态 + 写 outbox；由独立 consumer 调 `ledger.Transfer` 入账（与 hunter-pay「回调 → MQ → 入账」一致，且解耦）。
- 平台账户：`GetOrCreateAccountByOwner("platform", currency)` 作为**系统账户**（`config.PlatformAccountOwner`），入账 Transfer 的 from；系统账户允许透支/初始大额余额（实现时 seed 或放开余额校验）。

## 8. 提现 Saga（补偿模式）

```
前端              payment 模块                     NATS                  支付宝
  │ POST /withdraw │                               │                     │
  ├───────────────►│ CreateWithdrawalCommand       │                     │
  │                │  UoW: ledger.Freeze(用户, 金额, 幂等=out_biz_no)     │
  │                │  UoW: 保存 Withdrawal(FROZEN) │                     │
  │                │  UoW: 写 outbox(WithdrawalRequested)               │
  │                │       (relay) ─────────────────────────────────────►│ payment.cmd.withdrawal.requested
  │                │                         WithdrawalConsumer          │
  │                │                          AlipayPort.TransferToAlipayAccount
  │                │                          ─────────────────────────────────►│ 打款
  │                │                ◄─────────────────────────────────────┤ 结果
  │                │  成功: UoW ledger.WithdrawFromFrozen(用户, 金额, 幂等) + Withdrawal(SUCCESS)
  │                │  失败: UoW ledger.Unfreeze(用户, 金额, 幂等) 补偿解冻 + Withdrawal(FAILED)
```

要点：
- 冻结 = 资金「预留」；打款成功 = 确认扣减（`WithdrawFromFrozen` 永久减少冻结余额）；打款失败 = 补偿解冻（`Unfreeze` 恢复可用余额）。
- `out_biz_no` 同时作为支付宝幂等键与 ledger `IdempotencyKey`，consumer 重复投递安全。

## 9. MQ / 事件主题（NATS JetStream）

- 新增 stream：`PAYMENT_EVENTS`（或并入既有事件流，按 `.env` 配置）
- subjects：
  - `payment.evt.deposit.success`   → `DepositSuccessConsumer`（调 ledger 入账）
  - `payment.cmd.withdrawal.requested` → `WithdrawalConsumer`（调支付宝打款 + Saga 补偿）
  - `payment.evt.withdrawal.completed` / `payment.evt.withdrawal.failed`（可选，供通知/审计）
- 事件经 `payments_outbox` 表 + 本模块 relay 异步发布（与 `deposit_outbox` 同机制，`published_at IS NULL` 待发 + Nats-Msg-Id 去重）。

## 10. 数据库迁移

- `000018_create_payments_table.sql`：`payments(id BIGSERIAL PK, user_id, amount_cents, currency, status, out_trade_no VARCHAR UNIQUE, qr_code_url, alipay_trade_no, version INT, created_at, updated_at)`
- `000019_create_withdrawals_table.sql`：`withdrawals(id, user_id, ledger_account_id, alipay_account, alipay_real_name, amount_cents, currency, status, out_biz_no VARCHAR UNIQUE, frozen_op_id, alipay_order_id, fail_reason, version, created_at, updated_at)`
- `000020_create_payments_outbox_table.sql`：同 `deposit_outbox` 结构（`payments_outbox`）

## 11. 配置（扩展 `config`）

新增 `config.Alipay`：
- `ALIPAY_APP_ID`、`ALIPAY_APP_PRIVATE_KEY`、`ALIPAY_PUBLIC_KEY`
- `ALIPAY_GATEWAY`（生产 `https://openapi.alipay.com/gateway.do` / 沙箱 `https://openapi.alipaydev.com/gateway.do`）
- `ALIPAY_NOTIFY_BASE_URL`（公网回调基址，拼接 `/api/v1/payment/alipay/notify`）
- `PAYMENT_PROVIDER`：`alipay` | `mock`（默认 `mock`，无凭证可编译/单测/本地运行）
- `PLATFORM_ACCOUNT_OWNER`：系统账户 owner 标识（默认 `platform`）

## 12. HTTP 端点（注册到 `payment.Module` 路由）

- `POST   /api/v1/payment/deposit`
- `GET    /api/v1/payment/deposit/:id`
- `POST   /api/v1/payment/withdraw`
- `GET    /api/v1/payment/withdraw/:id`
- `POST   /api/v1/payment/alipay/notify`（公网回调，验签）

`cmd/all.go`、`cmd/auction.go` 增加 `payment.Module`，并在 fx 生命周期 `OnStart` 启动 relay 与两个 consumer 订阅。

## 13. 实施步骤（建议顺序）

1. 迁移 `000018/000019/000020` + `sqlc.yaml` 追加 payment 查询 + `make sqlc`
2. `config.Alipay` + `.env.example` 补充
3. domain（`payment_model`/`withdrawal_model`/enum/errs/event）
4. ports（`payment_repository`/`payment_uow`/`alipay_port`）
5. infra：repository(pg) / uow / alipay(sdk+mock+factory) / event envelope / outbox(relay)
6. application：command（create_deposit/create_withdrawal/alipay_notify）+ query + service（两个 consumer）
7. infra/http：handler + router + dto + notify handler；`payment.Module` 装配
8. `cmd` 入口接入 + `RegisterPaymentRoutes` + 启动 relay/consumer
9. `make update-mocks`；单测（mock 支付宝 + mock ledger）
10. 验证门禁：`build/vet/test/lint/nilaway` + testcontainers 集成测试

## 14. 验证

- 单测：mock `AlipayPort` + mock `LedgerRepository` 覆盖充值成功/失败、提现成功/打款失败补偿。
- 集成（testcontainers 真实 PG）：充值回调 → outbox → consumer → ledger 余额增加；提现 → 冻结 → consumer 打款失败 → 解冻（用 mock 支付宝使打款失败路径可达）。
- 全部门禁保持：`go build/vet/test/lint/nilaway` 全绿。

## 15. 风险与待决

- **回调公网可达**：支付宝回调需公网 URL；开发用 ngrok/cloudflared 隧道，`ALIPAY_NOTIFY_BASE_URL` 配置。代码侧保证验签与重放防护。
- **平台账户余额来源**：系统账户需预置初始余额或放开透支校验（实现时确定，见下方选型）。
- **真实对接需沙箱凭证**：默认 `mock` provider 保证本地/CI 可跑；接入真实支付宝需沙箱 APP_ID/密钥。
- **当面付 vs 转账范围**：本期仅做这两类（与 hunter-pay 一致）；后续可扩展退款、对账等。
