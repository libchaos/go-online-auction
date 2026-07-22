# 设计文档：保证金（押金 / Escrow）模块

> 适用工程：`go-online-auction`（Go + React 实时在线拍卖，六边形架构 + CQRS + DDD + Uber Fx + NATS JetStream + PostgreSQL/sqlc-pgx）
> 关联设计：`docs/trading-mode-strategy-design.md`（多交易模式策略）
> 目标：在现有架构内，新增一个与拍卖模块**解耦**的保证金模块，用于「资格审查闸门 + 结算（释放/抵扣/罚没）」。

---

## 1. 背景与现状

### 1.1 现状

当前任意通过鉴权的 `bidder`（`authn.Claims.Role == RoleBidder`）都可以对活跃拍卖出价，出价链路为：

- HTTP 层 `auction_handler.PlaceBid` 从上下文取 `claims.UserID`，调用 `placeBidCommand.Execute`；
- 命令仅把 `ports.BidCommand` 发布到 JetStream 命令流 `auction.cmd.bid.<auctionID>`，立即返回 `{status:"accepted"}`；
- 真正领域处理在 `BidProcessor.ProcessBid`（`infra/messaging/bid_processor.go`）内：加载 `AuctionModel` → `resolver.ForMode(...)` → `auction.PlaceBid(bid.Amount())` → `BidRepository.Create` → 同一事务写 `event_outbox` → `commitAndDispatch`。

`AuctionModel.PlaceBid`（`internal/modules/auction/domain/model/auction_model.go`）目前仅做英式硬编码校验（`state==active`、`!expired`、`amount > highest`），**没有任何资金/押金维度的约束**。拍卖结束时 `Close`/`DetermineWinner` 只算出赢家并发布 `AuctionEndedEvent`，**不触碰任何押金结算**。

### 1.2 问题

1. **无「入场门槛」**：任何 bidder 都能随意出价，存在恶意抬价、悔拍风险。
2. **无结算联动**：中标后没有机制把押金抵扣货款，也未中标者没有自动释放路径；平台缺乏违约惩罚手段。
3. **资金动作缺失抽象**：系统当前不持有任何「锁资/释放/扣款」能力，难以对接真实支付/钱包。

### 1.3 目标

| 维度 | 目标 |
|---|---|
| 资格审查 | 竞拍前必须持有足额押金，否则出价被拒（同步即时反馈 + 异步强校验双层） |
| 结算 | 拍卖结束：中标者押金抵扣货款；未中标者押金自动释放；违约者押金罚没 |
| 架构一致 | 复用六边形 + CQRS + Fx；模块间通过**端口 + 领域事件**解耦，auction 不直接 import deposit |
| 支付抽象 | 资金流转的**最终仲裁者**是模块内 `ledger` 子账本（`Freeze`/`Unfreeze`/`WithdrawFromFrozen` 原子 + 幂等键）；`PaymentPort` 降为可选外部动作适配器，v1 用 `MockPaymentAdapter` 占位，与账本解耦对账 |
| 可观测 | 押金状态变更经 outbox → JetStream → WebSocket 推送给对应用户 |

---

## 2. 关键假设与设计决策

> 这些决策若与你的预期不符，直接告诉我即可调整；下文均以此为基准展开。

1. **资金流转的最终仲裁者是模块内 `ledger` 子账本（参考 hunter-pay 的 SimpleBank）**。新增 `ledger` 子模块，持有 `ledger_accounts` / `ledger_entries` / `ledger_transfers` / `ledger_operations` 四张表，提供 `CreateAccount` / `Transfer` / `Freeze` / `Unfreeze` / `WithdrawFromFrozen` 五个**原子操作**（单事务 + `SELECT … FOR UPDATE` + 乐观 `version` 校验 + 幂等键唯一约束），作为扣款 / 冻结 / 释放的唯一真相源。`deposits` 表仍是押金意图账本（记录生命周期状态与金额），其内部资金动作（冻结 / 释放 / 抵扣 / 罚没）一律改写为调用 `ledger` 操作，不再直接走 `PaymentPort`。`PaymentPort` 保留为可选外部动作适配器（v1 `MockPaymentAdapter` 永远成功），未来接真实 PSP 时与账本 `external_ref` 对账，与核心资金裁决解耦。
2. **押金以 `(user_id, auction_id)` 为粒度**（一场拍卖一笔押金）。账户级「通用信用押金」作为未来扩展，本文档不展开。
3. **金额以 `uint64` 最小货币单位（分）存储**，币种单列 `currency varchar NOT NULL DEFAULT 'CNY'`。沿用现有 `MoneyModel` 风格（值对象、不可变、私有字段 + getter）。
4. **跨模块解耦两路**：
   - *资格审查*——auction 模块通过端口 `depositports.DepositGuard.EnsureEligible(...)` 同步预校验（deposit 模块实现并注册进 Fx）；
   - *结算*——deposit 模块**订阅 `AuctionEndedEvent`**（事件驱动）完成释放/抵扣，auction 不反向依赖 deposit。
5. **遵循 `.agent` 规范**：枚举 `Enum{Name}{Value}` + `validate{Name}Enum` 映射 + `errors.New`；领域错误集中在 `domain/errs/errs.go`；值对象不可变按值返回；测试用 `testify` `suite` + AAA 注释。

---

## 3. 保证金状态枚举

路径：`internal/modules/deposit/domain/enum/deposit_status_enum.go`。完全照抄现有 `trading_mode_enum.go` 范式。

```go
package enum

import "errors"

const (
    EnumDepositStatusPending   string = "pending"   // 已创建，外部锁资未确认
    EnumDepositStatusHeld      string = "held"      // 资金已锁定，具备出价资格
    EnumDepositStatusReleased  string = "released"  // 已释放回用户（未中标/取消/退款）
    EnumDepositStatusApplied   string = "applied"   // 已抵扣中标货款
    EnumDepositStatusForfeited string = "forfeited" // 已罚没（中标者违约）
)

type DepositStatusEnum struct{ value string }

func NewDepositStatusEnum(value string) (DepositStatusEnum, error) {
    if err := validateDepositStatusEnum(value); err != nil {
        return DepositStatusEnum{}, err
    }
    return DepositStatusEnum{value: value}, nil
}

func (e *DepositStatusEnum) String() string { return e.value }

func validateDepositStatusEnum(value string) error {
    allowed := map[string]struct{}{
        EnumDepositStatusPending:   {},
        EnumDepositStatusHeld:      {},
        EnumDepositStatusReleased:  {},
        EnumDepositStatusApplied:   {},
        EnumDepositStatusForfeited: {},
    }
    if _, ok := allowed[value]; !ok {
        return errors.New("invalid deposit status: " + value)
    }
    return nil
}
```

状态迁移表（领域方法负责守卫非法迁移）：

| 当前 \ 事件 | ConfirmHold | Release | ApplyToWinning | Forfeit | Cancel |
|---|---|---|---|---|---|
| `pending` | → `held` | — | — | — | → `released` |
| `held` | — | → `released` | → `applied` | → `forfeited` | — |
| `released`/`applied`/`forfeited` | 终态（拒绝） | 终态 | 终态 | 终态 | 终态 |

---

## 4. 领域模型（DepositModel）

### 4.1 聚合结构

路径：`internal/modules/deposit/domain/model/deposit_model.go`。

```go
type DepositModel struct {
    id         uint64
    userID     uint64
    auctionID  uint64
    amount     MoneyModel
    currency   string
    status     enum.DepositStatusEnum
    externalRef string   // 外部 PSP 交易引用，锁资成功后回填
    reference  string   // 幂等键 / 业务引用
    version    uint64
    createdAt  time.Time
    updatedAt  time.Time
}
```

构造：`NewDeposit(userID, auctionID uint64, amount MoneyModel, currency, reference string) (DepositModel, error)` —— 校验 `userID>0`、`auctionID>0`、`amount>0`、`currency` 非空，初始状态 `pending`，`version=1`。

### 4.2 状态机与领域行为

每个迁移方法先守卫当前状态，非法迁移返回 `errs.ErrInvalidDepositTransition`；成功则更新 `status` 并 `version++`、`updatedAt=now`。

```go
func (d *DepositModel) ConfirmHold(externalRef string) error // pending -> held
func (d *DepositModel) Release() error                       // held -> released
func (d *DepositModel) ApplyToWinning() error                // held -> applied
func (d *DepositModel) Forfeit() error                       // held -> forfeited
func (d *DepositModel) Cancel() error                        // pending -> released

func (d *DepositModel) IsEligible(required MoneyModel) bool  // status==held && amount>=required
func (d *DepositModel) Status() enum.DepositStatusEnum
func (d *DepositModel) Amount() MoneyModel
```

领域错误（集中定义于 `internal/modules/deposit/domain/errs/errs.go`，纯 `errors.New`）：

```go
var (
    ErrDepositRequired        = errors.New("a deposit is required to bid on this auction")
    ErrDepositInsufficient    = errors.New("deposit amount is insufficient for this auction")
    ErrDepositNotHeld         = errors.New("no held deposit found for this user and auction")
    ErrDepositAlreadyHeld     = errors.New("a deposit is already held for this user and auction")
    ErrInvalidDepositTransition = errors.New("invalid deposit status transition")
    ErrDepositExternalFailure = errors.New("external payment provider failed to hold funds")
)
```

### 4.3 MoneyModel 值对象

路径：`internal/modules/deposit/domain/model/money_model.go`。沿用 auction 的同名类型（`uint64` 分、不可变、私有字段 + getter），并补充本模块需要的 `Subtract`/`Compare`/`IsZero`：

```go
func NewMoneyModel(amountInCents uint64) MoneyModel
func (m MoneyModel) AmountInCents() uint64
func (m MoneyModel) IsGreaterThan(other MoneyModel) bool
func (m MoneyModel) IsGreaterThanOrEqual(other MoneyModel) bool
func (m MoneyModel) Subtract(other MoneyModel) (MoneyModel, error) // 不足额返回 Err...
```

> 说明：当前 auction 的 `MoneyModel` 缺 `Subtract/Compare`。若多个模块强共享，建议抽到 `internal/shared/modules/money`；v1 在 deposit 内自含一份，避免跨模块重构。

---

## 5. 端口（Ports）

### 5.1 DepositRepository

`internal/modules/deposit/ports/deposit_repository.go`：

```go
type DepositRepository interface {
    Save(ctx context.Context, d model.DepositModel) (model.DepositModel, error)
    FindByID(ctx context.Context, id uint64) (model.DepositModel, error)
    FindByUserAndAuction(ctx context.Context, userID, auctionID uint64) (model.DepositModel, error)
    ListByUser(ctx context.Context, userID uint64) ([]model.DepositModel, error)
    ListHeldByAuction(ctx context.Context, auctionID uint64) ([]model.DepositModel, error)
    Update(ctx context.Context, d model.DepositModel) (model.DepositModel, error)
}
```

### 5.2 PaymentPort（外部资金动作抽象 —— 六边形关键端口）

`internal/modules/deposit/ports/payment_port.go`。领域与 HTTP 层只依赖此接口，不感知真实 PSP。

```go
type PaymentPort interface {
    Hold(ctx context.Context, userID uint64, amount model.MoneyModel, currency, reference string) (externalRef string, err error)
    Release(ctx context.Context, externalRef string) error
    Capture(ctx context.Context, externalRef string, amount model.MoneyModel) error // 抵扣部分/全部
    Forfeit(ctx context.Context, externalRef string) error
}
```

适配器：`internal/modules/deposit/infra/payment/mock_payment_adapter.go`（`MockPaymentAdapter`，`Hold` 直接返回 `reference` 作为 `externalRef`，其余空实现成功）—— 供 dev/test/Fx 默认注入；真实 PSP 适配器后续新增。

> **架构演进**：自引入 `ledger` 子模块后，`PaymentPort` 不再是资金流转的仲裁者。`CreateDeposit` / `Release` / `Apply` / `Forfeit` / `Cancel` 五个命令的内部资金动作已改为调用 `ledger`（`Freeze` / `Unfreeze` / `WithdrawFromFrozen`），`PaymentPort` 仅作为可选的外部资金动作适配层存在，未来接真实 PSP 时与账本 `external_ref` 对账。

### 5.3 DepositGuard（资格审查，供 auction 模块调用）

`internal/modules/deposit/ports/deposit_guard.go`：

```go
type DepositGuard interface {
    EnsureEligible(ctx context.Context, userID, auctionID uint64) error
}
```

实现读取拍卖要求的押金额（来自 `auctions.deposit_required`/`deposit_amount_in_cents`）→ `DepositRepository.FindByUserAndAuction` → `DepositModel.IsEligible(required)`。不满足返回 `ErrDepositRequired` / `ErrDepositInsufficient`。deposit 模块在 `module.go` 把实现 `fx.As(new(depositports.DepositGuard))` 注册，auction 模块通过 Fx 注入使用。

### 5.4 事务性 Outbox 与 UoW

现有 `event_outbox` 表与 relay 位于 `auction` 模块内（`infra/outbox/relay.go`、`infra/event/envelope`）。建议**把 outbox relay + repository + envelope 下沉为 `internal/shared/modules/outbox`**（共享），deposit 与 auction 共用同一 `event_outbox` 表与 relay 进程；若暂不想动 auction，deposit 可先复制一份 outbox 实现到本模块（独立 `deposit_event_outbox` 表）。本文档按「下沉共享」给出设计。

---

### 5.5 Ledger 子模块（资金流转的最终仲裁者）

路径前缀：`internal/modules/ledger/`（顶层模块，domain / ports / application/service / infra{mapper,repository,uow,query,sqlcgen,http}）。参考 hunter-pay `SimpleBank`：以账户余额 + 冻结余额 + 不可变分录（entries）构成内部资金账本，所有跨账户资金移动都是**原子事务 + 幂等键**的单一操作。

**四张表（迁移 `000012`–`000015`）**

| 表 | 关键列 / 约束 | 作用 |
|---|---|---|
| `ledger_accounts` | `owner, balance, frozen_balance, currency, version`；`uq_ledger_accounts_owner_currency(owner, currency)` 部分唯一 | 账户主表，`balance>=0`、`frozen_balance>=0` 守卫；`version` 乐观锁 |
| `ledger_entries` | `account_id, amount(有符号), entry_type, operation_id` | 不可变资金流水（deposit/withdraw/transfer_out/in/freeze/unfreeze/withdraw_from_frozen/…） |
| `ledger_transfers` | `from_account_id, to_account_id, amount, idempotency_key UNIQUE` | `Transfer` 幂等键唯一约束 |
| `ledger_operations` | `account_id, counterparty_account_id, operation_type(freeze/unfreeze/withdraw_from_frozen), amount, idempotency_key UNIQUE, status, reference, description` | `Freeze`/`Unfreeze`/`WithdrawFromFrozen` 幂等键唯一约束 + 审计 |

**五个原子操作（接口 `ledger/ports/ledger_repository.go`）**

- `CreateAccount(owner, currency)` / `GetOrCreateAccountByOwner`：`owner+currency` 唯一，冲突则回退查询。
- `Transfer(input)`：先按 `idempotency_key` 查重 → `FOR UPDATE` 锁双账户 → 余额守卫 → 双账户余额变更（同 `version` 乐观校验）→ 两笔 `entry`（transfer_out / transfer_in）→ 写 `ledger_transfers`（唯一冲突则回查已存在记录返回）。
- `Freeze(input)`：幂等查重 → `FOR UPDATE` 锁账户 → 余额守卫 → `balance -= amount; frozen += amount` → 写 `ledger_operations(freeze)` → 写 `entry(freeze)`。
- `Unfreeze(input)`：镜像操作，`frozen -= amount; balance += amount`。
- `WithdrawFromFrozen(input)`：需 `counterparty_account_id != 0`；`frozen -= amount`（本账户）→ `balance += amount`（对手账户）→ 写 `ledger_operations(withdraw_from_frozen, counterparty_account_id)` + 两笔 `entry`（withdraw_from_frozen / withdraw_from_frozen_credit）。

**幂等性策略（DB 唯一约束为真相源）**：HTTP `Idempotency-Key` 头 → 缺省回退请求体字段 → 再回退 `uuid.NewString()`；该键透传到 ledger 操作的 `idempotency_key` 列。命令侧合成确定性键：如 `"deposit:"+id+":release"` / `:apply` / `:forfeit` / `:cancel`，保证同一押金的生命周期动作重复提交安全。仓库层采用「先查后插 + 唯一约束捕获回退（`isUniqueViolation`）」双保险。

**与 `DepositUnitOfWork` 同事务**：`DepositUnitOfWork` 扩展出 `LedgerRepository()`，ledger 仓库复用**同一个 `pgx.Tx`**；`CreateDeposit` 等命令在 UoW 内先调 ledger（`Freeze`/`Unfreeze`/`WithdrawFromFrozen`）再落 `deposits` 状态，整体原子提交 / 回滚。

---

## 6. 应用层（CQRS）

### 6.1 命令（写）

资金类操作需要即时反馈，**v1 采用同步命令**（HTTP handler → command → UoW → outbox），不走高延迟的 NATS 命令流；如团队坚持全异步，可改为 `deposit.cmd.*` 命令流。

- `CreateDepositCommand`：`{UserID, AuctionID, AmountInCents, Currency, IdempotencyKey}` → 在 UoW 内 `ledger.GetOrCreateAccountByOwner(owner=strconv(userID), currency)` → `ledger.Freeze(accountID, amount, idempotencyKey=reference, description)`（原子 + 幂等；失败直接透传 ledger 错误）→ `NewDeposit`(pending) → `ConfirmHold(op.IdempotencyKey())` → `Save` → 写 `DepositHeldEvent` outbox。返回 `{DepositID, Status:"held", AccountID}`。`UNIQUE(user_id, auction_id)` 防重复押金。
- `ReleaseDepositCommand`：`{DepositID}` → 加载 → 若原状态为 `held` 则 `ledger.Unfreeze(buyerAccountID, amount, key("release"))` → `Release` → `Save` → `DepositReleasedEvent`。
- `ApplyDepositCommand`：`{DepositID, CaptureAmountInCents}` → `ApplyToWinning` → `ledger.WithdrawFromFrozen(buyerAccountID, platformAccountID, captureAmount, key("apply"), desc)`（买家冻结余额扣减、平台账户余额增加）→ `Save` → `DepositAppliedEvent`。
- `ForfeitDepositCommand`：`{DepositID}` → `Forfeit` → `ledger.WithdrawFromFrozen(buyerAccountID, platformAccountID, amount, key("forfeit"), desc)`（罚没抵扣平台）→ `Save` → `DepositForfeitedEvent`。
- `CancelDepositCommand`：`{DepositID}` → `pending` 与 `held` 均可 `Cancel`；若原为 `held` 则 `ledger.Unfreeze(buyerAccountID, amount, key("cancel"))` 释放冻结 → `released`。

### 6.2 查询（读）

- `GetDepositQuery`（按 id）
- `ListDepositsByUserQuery`
- `ListHeldDepositsByAuctionQuery`（结算用）
- `GetEligibilityQuery`（封装 `DepositGuard.EnsureEligible` 供前端预检 UI）

### 6.3 领域事件与 JetStream 约定

路径：`internal/modules/deposit/domain/event/*_event.go`，基类沿用 `DomainEvent`（`eventID`=uuid、`timestamp`）。信封沿用 `envelope.Envelope`（`SchemaVersion=1`、`Data json.RawMessage`）。

| 事件 | EventType 常量 | subject |
|---|---|---|
| `DepositHeldEvent` | `deposit_held` | `deposit.evt.<userID>` |
| `DepositReleasedEvent` | `deposit_released` | `deposit.evt.<userID>` |
| `DepositAppliedEvent` | `deposit_applied` | `deposit.evt.<userID>` |
| `DepositForfeitedEvent` | `deposit_forfeited` | `deposit.evt.<userID>` |

新增 NATS 流/主题（`internal/shared/modules/nats/streams.go`）：

```go
StreamDepositEvents = "DEPOSIT_EVENTS"   // LimitsPolicy，可回放
SubjectDepositEvents = "deposit.evt.*"
```

发布：命令在 UoW 内 `OutboxRepository().Save(ctx, envelope)` → 共享 relay 轮询 `event_outbox`（`published_at IS NULL`）→ `js.Publish(ctx, evt.Subject, evt.Payload, jetstream.WithMsgID(evt.EventID))` 去重。

---

## 7. 与现有架构的集成

### 7.1 资格审查闸门（PlaceBid 两层）

- **同步预校验（即时反馈）**：在 `auction_handler.PlaceBid` 取到 `claims.UserID` 后、调用 `placeBidCommand.Execute` 之前，注入 `depositports.DepositGuard.EnsureEligible(ctx, userID, auctionID)`；不满足直接 4xx（映射 `ErrDepositRequired`/`ErrDepositInsufficient`）。
- **异步强校验（最终一致）**：在 `BidProcessor.ProcessBid` 内、`auction.PlaceBid(bid.Amount())` 之前调用同一端口；失败返回 `errs.ErrDepositNotHeld`，并在 `IsPermanentBidError` 中登记为**永久错误**（进 DLQ，不重试）。

### 7.2 结算（事件驱动，模块解耦）

deposit 模块启动一个 JetStream 消费者订阅 `SubjectEvents`（`auction.evt.*`），过滤 `auction_ended` 事件，解析出 `auctionID` 与 `winner_user_id`，执行：

1. `ListHeldByAuction(auctionID)` 取该拍卖所有 `held` 押金；
2. 对 `winner_user_id` 的那笔 → `ApplyDepositCommand`（抵扣中标货款，必要时 `PaymentPort.Capture`）；
3. 其余全部 → `ReleaseDepositCommand`（释放回用户）。

auction 模块**不 import deposit**，仅通过已发布的 `AuctionEndedEvent` 解耦协作。

### 7.3 WebSocket 广播

沿用 `websocket.Hub` + 订阅注册表模式：新增 `DepositSubscriberRegistry`（按 `userID` 订阅），消费者把 `deposit.evt.<userID>` 推给该用户的所有 WS 连接。前端据此刷新「我的押金」状态。

### 7.4 Fx 装配

`internal/modules/deposit/module.go`：

```go
var Module = fx.Module("deposit",
    fx.Provide(fx.Annotate(NewPostgresDepositRepository, fx.As(new(ports.DepositRepository)))),
    fx.Provide(fx.Annotate(NewMockPaymentAdapter, fx.As(new(ports.PaymentPort)))),
    fx.Provide(fx.Annotate(NewDepositGuard, fx.As(new(ports.DepositGuard)))),
    fx.Provide(NewCreateDepositCommand),
    fx.Provide(NewReleaseDepositCommand),
    fx.Provide(NewApplyDepositCommand),
    fx.Provide(NewForfeitDepositCommand),
    fx.Provide(NewCancelDepositCommand),
    fx.Provide(NewDepositHandler),
    fx.Invoke(RegisterDepositRoutes),
    fx.Invoke(RegisterDepositEventConsumer), // 订阅 auction_ended + 发布 deposit.evt.*
)
```

`main` 中调用 `deposit.RegisterDepositRoutes(server, handler, middleware)` 挂载路由；`auction` 模块通过 Fx 自动拿到 `depositports.DepositGuard` 实现。

---

## 8. 数据库迁移

> 风格：单个 `.sql` 文件 + `-- +goose Up` / `-- +goose Down`；金额 `bigint`（分）；状态 `varchar` + `CHECK`；id `BIGSERIAL`；乐观锁 `version BIGINT`；由 `migrations/embed.go` 的 `//go:embed *.sql` 嵌入。

**`migrations/000009_create_deposits_table.sql`**

```sql
-- +goose Up

CREATE TABLE deposits (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL,
    auction_id       BIGINT NOT NULL,
    amount_in_cents  BIGINT NOT NULL,
    currency         VARCHAR NOT NULL DEFAULT 'CNY',
    status           VARCHAR NOT NULL DEFAULT 'pending',
    external_ref     VARCHAR,
    reference        VARCHAR,
    version          BIGINT NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_deposits_status CHECK (status IN ('pending','held','released','applied','forfeited')),
    CONSTRAINT uq_deposits_user_auction UNIQUE (user_id, auction_id)
);

CREATE INDEX idx_deposits_user_id ON deposits(user_id);
CREATE INDEX idx_deposits_auction_id ON deposits(auction_id);
CREATE INDEX idx_deposits_status ON deposits(status);

-- +goose Down

DROP TABLE IF EXISTS deposits;
```

**`migrations/000010_add_auction_deposit_config.sql`**（拍卖维度的押金要求，供 `DepositGuard` 读取）

```sql
-- +goose Up

ALTER TABLE auctions ADD COLUMN deposit_required boolean NOT NULL DEFAULT false;
ALTER TABLE auctions ADD COLUMN deposit_amount_in_cents bigint NOT NULL DEFAULT 0;

-- +goose Down

ALTER TABLE auctions DROP COLUMN IF EXISTS deposit_required;
ALTER TABLE auctions DROP COLUMN IF EXISTS deposit_amount_in_cents;
```

---

## 9. 如何扩展

- **新增支付方式**：实现 `PaymentPort` 新适配器（如 Stripe/支付宝），在 `deposit/module.go` 把 `fx.As(new(ports.PaymentPort))` 指向它；领域与命令零改动。
- **账户级通用押金**：`auction_id` 改为可空 `BIGINT`（允许 `NULL`），`UNIQUE(user_id, auction_id)` 调整为部分唯一索引；`DepositGuard` 优先用拍卖专属押金，回退到账户级押金。
- **多币种**：`currency` 已单列；金额比较需同币种守卫（领域方法加币种相等校验）。
- **违约自动罚没**：新增调度器订阅拍卖支付截止时间，超时未付则 `ForfeitDepositCommand`。

---

## 10. 测试策略

- **领域单测**（`domain/model/deposit_model_test.go`）：状态机每个迁移的合法/非法路径、守卫错误、`IsEligible` 边界。
- **命令/查询测试**：`mock_DepositRepository` + `mock_LedgerRepository`（`.Maybe()` 处理可选调用），验证 UoW 内状态流转、ledger 原子动作调用与 outbox 写入；`Cancel` 现允许从 `held` 释放冻结。
- **Ledger 仓库测试**：`ledger_repository` 的原子性（并发 `Transfer` / `Freeze` 余额守卫）、`version` 乐观锁冲突、幂等键唯一约束（重复提交安全返回首笔）、`WithdrawFromFrozen` 必须带 `counterparty_account_id`。
- **Ledger 服务 / HTTP 测试**：`LedgerAccountService` 经 UoW 编排；handler 校验 `Idempotency-Key` 透传与鉴权。
- **集成测试**（`tests/integration`）：起 Postgres + NATS，跑通 `CreateDeposit → Hold → DepositHeldEvent → WebSocket 收到`；以及 `AuctionEndedEvent → 释放非中标者 / 应用中标者` 的端到端结算。
- **并发**：`version` 乐观锁冲突重试；结算竞态（拍卖结束瞬间多笔押金）用 `ListHeldByAuction` 事务内统一处理。

---

## 11. 风险与权衡

| 风险 | 说明 / 缓解 |
|---|---|
| 账本 vs 资金来源 | `ledger` 子模块是资金流转的**最终仲裁者**（账户余额 + 冻结余额 + 不可变分录），真实外部 PSP 经 `PaymentPort` 仅作可选对账层；内部资金动作一律走 ledger，避免「意图账本」与「真实余额」双写不一致 |
| 外部锁资迟到确认 | `Hold` 异步回执可能晚于 `Cancel`；以 `reference`/`external_ref` 做幂等对账，迟到确认落袋为 `held` 或安全释放 |
| PSP 不可用 | 押金卡在 `pending`；需对账/重试任务，不阻断拍卖主流程 |
| 跨模块耦合 | 结算走事件（auction 不反向依赖 deposit）；资格审查走 `DepositGuard` 端口（deposit 实现、auction 注入），保持单向依赖 |
| outbox 共享改造 | 把 relay/envelope 下沉到 `internal/shared/modules/outbox` 有少量 auction 重构成本，但避免双份 outbox 实现 |
| 结算竞态 | 拍卖结束与押金释放/抵扣并发；用 UoW + 乐观锁 + 事件幂等（JetStream `Nats-Msg-Id`）保证恰好一次 |

---

## 12. 实施清单（分阶段）

- **Phase 1（领域）**：`DepositStatusEnum` + `DepositModel` 状态机 + `errs` + `MoneyModel`（含 `Subtract`/`Compare`）。
- **Phase 2（持久化）**：迁移 `000009`/`000010` + `DepositRepository`/`mapper` + deposit `UnitOfWork`（含 Outbox）。
- **Phase 3（支付抽象）**：`PaymentPort` + `MockPaymentAdapter` + Fx 注册（现为可选外部动作适配层）。
- **Phase 4（CQRS 写/读）**：5 个命令 + 4 个查询 + 4 个领域事件（信封/outbox/JetStream `DEPOSIT_EVENTS`/`deposit.evt.<userID>`）。
- **Phase 9（资金账本）**：`ledger` 子模块 —— 迁移 `000012`–`000015` + sqlc + domain（Account/Operation/Entry/Transfer 模型 + 枚举 + errs）+ `LedgerRepository`（原子 `Transfer`/`Freeze`/`Unfreeze`/`WithdrawFromFrozen` + 幂等键）+ UoW（同事务挂入 `DepositUnitOfWork`）+ `LedgerAccountService` + HTTP 路由（`/api/v1/ledger/*`）+ 5 个命令改写为调用 ledger。
- **Phase 5（HTTP/WS）**：`deposit` 路由 + DTO + `bidder` 角色鉴权 + `DepositSubscriberRegistry` 推送。
- **Phase 6（资格审查）**：`DepositGuard` + `auction_handler.PlaceBid` 同步预校验 + `BidProcessor` 异步强校验（`ErrDepositNotHeld` 入 `IsPermanentBidError`）。
- **Phase 7（结算）**：`AuctionEndedEvent` 消费者 → 释放非中标者 / 应用中标者；违约罚没调度器（可选）。
- **Phase 8（质量）**：领域/命令/集成测试 + README 补充保证金说明。

---

## 附：关键文件索引

| 文件 | 角色 |
|---|---|
| `internal/modules/deposit/domain/enum/deposit_status_enum.go` | **新增** 押金状态枚举 |
| `internal/modules/deposit/domain/errs/errs.go` | **新增** 领域错误 |
| `internal/modules/deposit/domain/model/deposit_model.go` | **新增** 押金聚合 + 状态机 |
| `internal/modules/deposit/domain/model/money_model.go` | **新增** 金额值对象（含 Subtract/Compare） |
| `internal/modules/deposit/domain/event/*_event.go` | **新增** 4 个押金事件 |
| `internal/modules/deposit/ports/{deposit_repository,payment_port,deposit_guard}.go` | **新增** 端口（`PaymentPort` 现为可选外部动作适配层） |
| `internal/modules/ledger/domain/{model,enum,errs}` | **新增** 账本领域（Account/Operation/Entry/Transfer 模型 + 枚举 + 错误） |
| `internal/modules/ledger/ports/{ledger_repository,ledger_uow,ledger_account_service}.go` | **新增** 账本端口（原子操作 + 同事务 UoW + 服务接口） |
| `internal/modules/ledger/application/service/ledger_account_service.go` | **新增** 账本服务（`runTx` 编排） |
| `internal/modules/ledger/infra/{mapper,repository,uow,query,http}` | **新增** 账本适配层（原子仓库 + 同事务 UoW + HTTP `/api/v1/ledger/*`） |
| `internal/modules/deposit/application/command/*_command.go` | **改写** 5 个写命令内部资金动作改为调用 `ledger` |
| `internal/modules/deposit/application/query/*_query.go` | **新增** 4 个读查询 |
| `internal/modules/deposit/infra/{repository,mapper,payment,http/chi,websocket}` | **新增** 适配层 |
| `internal/modules/deposit/testmocks/mock_LedgerRepository.go` | **新增** 账本仓库 mock（命令测试用） |
| `internal/modules/deposit/module.go` | **新增** Fx 装配 + `RegisterDepositRoutes`/`RegisterDepositEventConsumer`/`RegisterLedgerRoutes` |
| `migrations/000009_create_deposits_table.sql` | **新增** 押金表 |
| `migrations/000010_add_auction_deposit_config.sql` | **新增** 拍卖押金要求列 |
| `migrations/000012_create_ledger_accounts_table.sql` | **新增** 账本账户表 |
| `migrations/000013_create_ledger_entries_table.sql` | **新增** 账本分录表 |
| `migrations/000014_create_ledger_transfers_table.sql` | **新增** 账本转账表（幂等键唯一） |
| `migrations/000015_create_ledger_operations_table.sql` | **新增** 账本操作表（幂等键唯一） |
| `internal/shared/modules/outbox` | **改造（建议）** 下沉共享 outbox relay/envelope |
| `internal/shared/modules/nats/streams.go` | **改造** 新增 `DEPOSIT_EVENTS` / `deposit.evt.*` |
| `internal/modules/auction/{application/command/place_bid_command,infra/http/chi/handler,infra/messaging/bid_processor}.go` | **改造** 接入 `DepositGuard` 闸门 |
