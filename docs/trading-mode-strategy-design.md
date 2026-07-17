# 设计文档：基于策略模式的多交易模式实现

> 适用工程：`go-online-auction`（Go + React 实时在线拍卖系统）
> 目标：在不修改现有竞价核心流程的前提下，支持英式、荷兰式、密封投标、维克里、一口价等多种交易模式，并符合本仓库已有的「六边形架构 + CQRS + Fx 依赖注入」风格。

---

## 1. 背景与现状

### 1.1 现状

当前出价业务规则被**硬编码**在领域模型里。见 `internal/modules/auction/domain/model/auction_model.go` 的 `PlaceBid`：

```go
func (a *AuctionModel) PlaceBid(amount MoneyModel) error {
    if a.state.String() != enum.EnumAuctionStateActive {
        return errs.ErrBidsOnlyOnActiveAuctions
    }
    if time.Now().UTC().After(a.endTime) {
        return errs.ErrAuctionExpired
    }
    // Validate bid amount against current highest  ← 这是“英式拍卖”规则
    if a.highestBidAmount == nil {
        if amount.AmountInCents() == 0 {
            return errs.ErrFirstBidMustBePositive
        }
    } else {
        currentHighest := NewMoneyModel(*a.highestBidAmount)
        if !amount.IsGreaterThan(currentHighest) {
            return errs.ErrBidMustExceedHighest   // ← 仅英式成立
        }
    }
    amountInCents := amount.AmountInCents()
    a.highestBidAmount = &amountInCents
    a.version++
    a.updatedAt = time.Now().UTC()
    return nil
}
```

异步消费者 `BidProcessor.ProcessBid`（`internal/modules/auction/infra/messaging/bid_processor.go:144`）调用 `auction.PlaceBid(money)`，并依赖 `IsPermanentBidError` 把业务错误路由到 DLQ。

### 1.2 问题

- 不同交易模式的核心规则差异极大（升价 / 降价 / 密封 / 第二价 / 一口价），若继续在 `PlaceBid` 里用 `if/switch` 堆叠，会迅速腐化域模型。
- `IsPermanentBidError`、`Close`、WebSocket 广播都隐含「英式」假设，扩展时需要在多处改同一逻辑。
- 违反了**开闭原则**：每新增一种模式都要改既有文件。

### 1.3 目标

| 维度 | 目标 |
|---|---|
| 可扩展 | 新增交易模式只需新增一个策略类 + 注册，不改既有流程（开闭原则） |
| 可测试 | 每种模式的规则可独立单元测试，复用仓库已有的 `*_test.go` 风格 |
| 可插拔 | 策略随拍卖实例绑定，由交易模式枚举驱动，运行时按枚举解析 |
| 架构一致 | 策略属于 domain 层；不引入新的基础设施；通过 Fx 装配；不破坏现有异步竞价链路 |

---

## 2. 交易模式枚举

沿用 `domain/enum/auction_state_enum.go` 的「常量 + 值对象」写法，新增 `domain/enum/trading_mode_enum.go`：

```go
const (
    EnumTradingModeEnglish    string = "english"     // 英式/升价（现行逻辑）
    EnumTradingModeDutch      string = "dutch"       // 荷兰式/降价
    EnumTradingModeSealedBid  string = "sealed_bid"  // 密封投标（盲拍）
    EnumTradingModeVickrey    string = "vickrey"     // 维克里/第二价密封
    EnumTradingModeFixedPrice string = "fixed_price" // 一口价
    EnumTradingModeEbayProxy  string = "ebay_proxy"  // 类 eBay 自动代理加价
)

// 注意：延时 / 防狙击（anti-snipe）不是一种“交易模式”，而是与模式正交的
// “关闭策略”，可叠加在任意模式之上（见第 5.7 节），因此不放入本枚举。

type TradingModeEnum struct{ value string }

func NewTradingModeEnum(value string) (TradingModeEnum, error) { /* 校验 + 返回，同 AuctionStateEnum */ }
func (e *TradingModeEnum) String() string { return e.value }
```

> 存量拍卖数据迁移时默认填充 `english`，保证向后兼容。

---

## 3. 策略接口设计

新增 `domain/strategy/strategy.go`。策略只**读取**领域模型状态（通过已导出的 getter），**不**直接修改未导出字段；状态变更交回 `AuctionModel` 的方法完成。

```go
package strategy

import (
    "auction/internal/modules/auction/domain/enum"
    "auction/internal/modules/auction/domain/model"
)

// Winner 是拍卖结束后的成交结果
type Winner struct {
    UserID     uint64
    PayAmount  model.MoneyModel // 实际支付价（维克里 ≠ 出价）
    BidID      *uint64
}

// TradingStrategy 封装一种交易模式的核心规则
type TradingStrategy interface {
    // Mode 返回该策略对应的交易模式
    Mode() enum.TradingModeEnum

    // ValidateBid 校验“此刻这笔出价是否被接受”。只读 auction 状态（getter）。
    ValidateBid(auction *model.AuctionModel, amount model.MoneyModel) error

    // SuggestNextPrice 返回下一笔有效出价的建议价格（用于前端提示与预校验）。
    // 不适用的模式（如密封投标）返回起拍价/保留价作为提示。
    SuggestNextPrice(auction *model.AuctionModel) (model.MoneyModel, error)

    // DetermineWinner 在拍卖关闭时计算获胜者与成交价。
    DetermineWinner(auction *model.AuctionModel, bids []model.BidModel) (Winner, error)

    // ShouldCloseOnAccept 出价被接受后是否立即结束拍卖（荷兰式/一口价=true，其余=false）。
    ShouldCloseOnAccept() bool
}
```

### 3.1 与领域模型的边界

- 策略只依赖 `AuctionModel` 的导出方法：`State()`、`HighestBidAmount()`、`CurrentPrice()`、`EndTime()`、`ReservePrice()`、`TradingMode()`、`AntiSnipeEnabled()`、`ExtensionWindowSec()`。
- 写入逻辑仍在 `AuctionModel`：`PlaceBid`（记录出价）、`Close()`、`ApplyWin(userID, amount)`、`MaybeExtendEndTime(now)`。
- 代理加价模式（`ebay_proxy`）还需读取 `BidModel.MaxAmount()`（出价人的授权上限），因此 `BidModel` 需新增 `maxAmount` 字段与 `MaxAmount()` getter（见第 4 节）。
- 这样策略保持「纯函数式」、易测，且域模型仍独占自身状态。

---

## 4. 领域模型改造

在 `auction_model.go` 增加交易模式相关字段与方法：

```go
type AuctionModel struct {
    id                uint64
    listingID         uint64
    tradingMode       enum.TradingModeEnum // 新增
    startingPrice     *uint64              // 新增：起拍价/一口价
    priceStep         *uint64              // 新增：最小加价/降价步长
    reservePrice      *uint64              // 新增：保留价（未达则流拍）
    currentPrice      *uint64              // 新增：当前有效价（荷兰式降价用）
    highestBidAmount  *uint64
    winnerUserID      *uint64              // 新增：成交用户
    winningBidAmount  *uint64              // 新增：成交价
    antiSnipeEnabled  bool                 // 新增：防狙击开关（正交，可叠加任意模式）
    extensionWindowSec int64               // 新增：防狙击窗口（秒），窗口内出价则顺延
    startTime         *time.Time
    endTime           time.Time
    state             enum.AuctionStateEnum
    version           uint64
    createdAt         time.Time
    updatedAt         time.Time
}

// BidModel 需同步新增 maxAmount（出价人授权上限），对外广播仍用 amount（隐藏 max）。
//   type BidModel struct {
//       ...
//       amount    MoneyModel
//       maxAmount *MoneyModel // 新增：代理加价模式下为出价人授权上限
//   }
//   func (b *BidModel) MaxAmount() *MoneyModel { return b.maxAmount }
```

### 4.1 `PlaceBid` 改造为委托策略

```go
func (a *AuctionModel) PlaceBid(amount MoneyModel) error {
    if a.state.String() != enum.EnumAuctionStateActive {
        return errs.ErrBidsOnlyOnActiveAuctions
    }
    if time.Now().UTC().After(a.endTime) {
        return errs.ErrAuctionExpired
    }
    s := strategy.Resolver().ForMode(a.tradingMode) // 按枚举解析策略
    if err := s.ValidateBid(a, amount); err != nil {
        return err
    }
    amountInCents := amount.AmountInCents()
    a.highestBidAmount = &amountInCents
    a.version++
    a.updatedAt = time.Now().UTC()
    return nil
}
```

> `strategy.Resolver()` 为包级单例（见第 6 节），保证 `RestoreAuctionModel` 重建的模型也能拿到策略，无需在持久化层注入策略对象。

### 4.2 `Start()` 初始化价格

```go
func (a *AuctionModel) Start() error {
    // ... 既有状态校验 ...
    now := time.Now().UTC()
    a.startTime = &now
    // 起拍：英式从起拍价开始；荷兰式从 startingPrice 开始递减
    if a.currentPrice == nil && a.startingPrice != nil {
        cp := *a.startingPrice
        a.currentPrice = &cp
    }
    a.state = activeState
    a.version++
    a.updatedAt = now
    return nil
}
```

### 4.3 `Close()` 计算成交结果

```go
func (a *AuctionModel) Close(bids []model.BidModel) error {
    if a.state.String() != enum.EnumAuctionStateActive {
        return errs.ErrAuctionCanOnlyCloseFromActive
    }
    s := strategy.Resolver().ForMode(a.tradingMode)
    winner, err := s.DetermineWinner(a, bids)
    if err != nil {
        return err
    }
    if winner.UserID != 0 {
        a.winnerUserID = &winner.UserID
        pay := winner.PayAmount.AmountInCents()
        a.winningBidAmount = &pay
    }
    a.state = closedState
    a.version++
    a.updatedAt = time.Now().UTC()
    return nil
}
```

### 4.4 `MaybeExtendEndTime()` 防狙击顺延

```go
// MaybeExtendEndTime 若启用防狙击且本次出价距结束不足窗口，则顺延 endTime。
// 不影响各模式的出价/成交规则，仅补充 Close()/CheckAndCloseIfExpired 的到期判定。
func (a *AuctionModel) MaybeExtendEndTime(now time.Time) {
    if !a.antiSnipeEnabled {
        return
    }
    window := time.Duration(a.extensionWindowSec) * time.Second
    if now.Before(a.endTime) && now.Add(window).After(a.endTime) {
        a.endTime = now.Add(window)
        a.version++
        a.updatedAt = now
    }
}
```

> 该方法由 `BidProcessor` 在「每次出价被成功接受后」调用（见第 7.4 节），与交易模式正交，可被英式/荷兰式/代理加价任意组合。

---

## 5. 各具体策略实现

> 本节含 5.1~5.5 五种 `TradingStrategy` 实现，以及 5.6 代理加价策略、5.7 防狙击关闭策略（正交机制，可叠加任意模式）。

### 5.1 英式拍卖 `EnglishAuctionStrategy`（现行逻辑平移）

| 方法 | 行为 |
|---|---|
| `ValidateBid` | 无出价则 `amount>0`；否则 `amount > currentHighest` |
| `SuggestNextPrice` | `highestBid + priceStep`（无步长则 +1） |
| `DetermineWinner` | 最高出价者以**其出价**成交；低于保留价则流拍 |
| `ShouldCloseOnAccept` | `false` |

### 5.2 荷兰式拍卖 `DutchAuctionStrategy`（降价）

当前价随时间递减：`currentPrice = max(reservePrice, startingPrice - priceStep * 已过去的时间段数)`。

| 方法 | 行为 |
|---|---|
| `ValidateBid` | `amount` 必须等于 `CurrentPrice()`（接受当前挂牌价）；否则 `ErrDutchBidMustMatchPrice` |
| `SuggestNextPrice` | 返回 `CurrentPrice()`（实时降价） |
| `DetermineWinner` | 第一个有效出价者以 `currentPrice` 成交 |
| `ShouldCloseOnAccept` | `true`（有人接受即结束） |

> 价格按「已过去时间段」计算，避免引入额外调度器；时钟一致性由服务节点时钟保证（见风险章节）。

### 5.3 密封投标 `SealedBidAuctionStrategy`（盲拍）

| 方法 | 行为 |
|---|---|
| `ValidateBid` | `amount>0` 即可，期间互不比较、隐藏 |
| `SuggestNextPrice` | 返回起拍价作为提示（无递增概念） |
| `DetermineWinner` | 结束时取最高出价者成交；低于保留价流拍 |
| `ShouldCloseOnAccept` | `false`（仅到时结束） |

### 5.4 维克里拍卖 `VickreyAuctionStrategy`（第二价密封）

| 方法 | 行为 |
|---|---|
| `ValidateBid` | 同密封投标（任意正数） |
| `SuggestNextPrice` | 同密封投标 |
| `DetermineWinner` | 最高出价者**获胜但按第二高出价支付**（不足则按保留价） |
| `ShouldCloseOnAccept` | `false` |

### 5.5 一口价 `FixedPriceAuctionStrategy`

| 方法 | 行为 |
|---|---|
| `ValidateBid` | `amount` 必须等于 `startingPrice`，否则 `ErrFixedPriceMismatch` |
| `SuggestNextPrice` | 返回 `startingPrice` |
| `DetermineWinner` | 第一个有效购买者以 `startingPrice` 成交 |
| `ShouldCloseOnAccept` | `true` |

### 5.6 自动代理加价模式 `EbayProxyAuctionStrategy`（类 eBay 代理出价）

**核心思想**：出价人提交的是**最高心理价位（max bid）**，系统在其授权范围内自动以「最小必要加价」代为出价，保证该出价人始终是当前最低领先者，直到被他人更高的 max 超越。本质仍是英式升价，但加了一层「代理机器人」。

**数据模型**
- `BidModel` 增加 `maxAmount *MoneyModel`（出价人授权上限），持久化列 `max_amount`。
- `amount` 仍是系统自动放置的**当前公开价**，对外广播 / `highestBidAmount` 用 `amount`（隐藏 max）。
- 领先者由 **max 最高** 决定，而非 `amount`。

**`ValidateBid(auction, amount)`**：此处 `amount` 语义为本次提交的 **max**。校验 `max > 当前公开最高价`（否则无法取得领先，返回 `ErrProxyMaxTooLow`）。

**代理出价解析算法**（纯函数，返回待写入的动作序列；持久化由 `BidProcessor` 在 UoW 内完成，避免 domain→ports 依赖倒置）：

```go
// ProxyAction 描述一条待持久化的代理出价
type ProxyAction struct {
    UserID   uint64
    Amount   model.MoneyModel // 系统自动放置的公开价
    MaxAmount model.MoneyModel // 出价人授权上限
}

// ResolveProxyBids 在 (新提交的 maxU) 进入后，计算收敛后的全部公开出价。
func (st *EbayProxyAuctionStrategy) ResolveProxyBids(
    auction *model.AuctionModel,
    existing []model.BidModel,   // 该拍卖全部已有出价（含 MaxAmount）
    newBidder uint64,
    newMax model.MoneyModel,
) ([]ProxyAction, error) {
    // 1. 按 user 聚合取每人最高 max（proxyMax[user]）
    // 2. 若 newBidder 已是领先者且 newMax <= 原 max：仅更新其上限，无新公开价
    // 3. 计算 newBidder 的公开出价 placed = min(currentPublicHighest + increment, newMax)
    //    - placed > newMax ⇒ 出局（不足以领先），领先者不变
    //    - 否则写入 placed，newBidder 成为领先者
    // 4. 反制循环：存在 other 使 proxyMax[other] > 当前公开最高价 placed：
    //      取 proxyMax 最大的 other，counter = min(placed + increment, proxyMax[other])
    //      若 counter <= proxyMax[other]：写入 counter，other 反超；重复步骤 4
    //    直至领先者的 proxyMax >= 所有其他人 proxyMax（必然收敛）
}
```

> **收敛性**：每次循环公开价严格递增（至少 +`priceStep`），且受所有人 max 上界约束，有限步内收敛。

**其余方法**

| 方法 | 行为 |
|---|---|
| `SuggestNextPrice` | `当前公开最高价 + priceStep`（新出价人的最小起拍 max） |
| `DetermineWinner` | **最高 max 者获胜**，支付价 = `次高 max + priceStep`（次高不存在则付起拍价），且不超过获胜者自身 max。这等价于**第二价格密封**结果——eBay 代理出价天然产生类维克里支付 |
| `ShouldCloseOnAccept` | `false`（仅到时结束） |

> 接入：`BidProcessor` 在 `PlaceBid` 成功后，对 `ebay_proxy` 模式调用 `ResolveProxyBids` 并批量 `Create`（见第 7.4 节）。

### 5.7 延时 / 防狙击关闭策略（正交关注点）

> **设计判断**：延时（反狙击 / soft close）与交易模式**正交**——它可叠加在英式、荷兰式、代理加价等任意模式之上，而非一种独立的「交易模式」。因此**不**放入 `TradingModeEnum`，而是建模为可独立开关的**关闭策略**，对任意模式复用，更契合开闭原则。若产品侧坚持把它当成一种「模式」对外展示，可再包一层策略门面，但底层复用同一机制（见第 4.4 节 `MaybeExtendEndTime`）。

**机制**：拍卖配置 `antiSnipeEnabled bool` 与 `extensionWindowSec int64`（如「最后 5 分钟内出价则延长 5 分钟」）。每次出价被成功接受后调用 `auction.MaybeExtendEndTime(time.Now().UTC())` 并 `Update`。

**组合示例**
- `english + antiSnipe`：最常见的标准英式拍卖带防狙击。
- `ebay_proxy + antiSnipe`：eBay 完整形态。
- `dutch + antiSnipe`：荷兰式降价拍卖同样可防狙击。

> 防狙击只改变 `endTime`，不改变各模式的出价 / 成交规则；它是对 `Close()` / `CheckAndCloseIfExpired` 的补充——顺延后需重新判定是否到期。

---

## 6. 策略工厂与 Fx 装配

### 6.1 `domain/strategy/resolver.go`

```go
type Resolver interface {
    ForMode(mode enum.TradingModeEnum) (TradingStrategy, error)
}

type mapResolver struct{ m map[string]TradingStrategy }

func NewResolver(strategies []TradingStrategy) Resolver {
    r := &mapResolver{m: make(map[string]TradingStrategy, len(strategies))}
    for _, s := range strategies {
        r.m[s.Mode().String()] = s
    }
    return r
}

func (r *mapResolver) ForMode(mode enum.TradingModeEnum) (TradingStrategy, error) {
    if s, ok := r.m[mode.String()]; ok {
        return s, nil
    }
    return nil, errs.ErrUnsupportedTradingMode
}
```

### 6.2 Fx 注册（`internal/modules/auction/module.go`）

利用 Fx 的 `fx.Group` 把多个策略收集为切片，再用 `fx.Annotate` 注册为 `Resolver`：

```go
// 各策略以 group 形式提供
fx.Provide(fx.Annotate(strategy.NewEnglishAuctionStrategy, fx.ResultTags(`group:"trading_strategy"`)))
fx.Provide(fx.Annotate(strategy.NewDutchAuctionStrategy,   fx.ResultTags(`group:"trading_strategy"`)))
fx.Provide(fx.Annotate(strategy.NewSealedBidAuctionStrategy, fx.ResultTags(`group:"trading_strategy"`)))
fx.Provide(fx.Annotate(strategy.NewVickreyAuctionStrategy,   fx.ResultTags(`group:"trading_strategy"`)))
fx.Provide(fx.Annotate(strategy.NewFixedPriceAuctionStrategy, fx.ResultTags(`group:"trading_strategy"`)))
fx.Provide(fx.Annotate(strategy.NewEbayProxyAuctionStrategy, fx.ResultTags(`group:"trading_strategy"`)))

// 聚合为 Resolver
fx.Provide(fx.Annotate(
    func(strategies strategy.Strategies `group:"trading_strategy"`) strategy.Resolver {
        return strategy.NewResolver(strategies)
    },
    fx.As(new(strategy.Resolver)),
))
```

> `strategy.Strategies` 为 `[]TradingStrategy` 的别名，配合 `group:"trading_strategy"` 标签收集。

---

## 7. 与现有架构的集成

### 7.1 domain 层（已述）
枚举、策略包、模型改造——均在第 2~4 节。

### 7.2 application 层

- **`CreateAuctionCommand`**：输入新增 `TradingMode string`、`StartingPrice`、`PriceStep`、`ReservePrice`；校验枚举后用 repository 持久化。
- **`CloseAuctionCommand` / `AuctionModel.CheckAndCloseIfExpired`**：关闭时传入该拍卖的全部出价（经 `BidRepository`），调用 `auction.Close(bids)` 计算 winner，再 dispatch 富化的 `AuctionEndedEvent`（携带 `WinnerUserID` / `WinningAmount`）。

### 7.3 ports & persistence

`AuctionRepository` / `BidRepository` 的 `FindByID`、`Create`、`Update` 需映射新增列：

| 模型字段 | 表列 |
|---|---|
| `tradingMode` | `trading_mode varchar not null default 'english'` |
| `startingPrice` | `starting_price bigint` |
| `priceStep` | `price_step bigint` |
| `reservePrice` | `reserve_price bigint` |
| `currentPrice` | `current_price bigint` |
| `winnerUserID` | `winner_user_id bigint null` |
| `winningBidAmount` | `winning_bid_amount bigint null` |
| `antiSnipeEnabled` | `anti_snipe_enabled boolean not null default false` |
| `extensionWindowSec` | `extension_window_sec bigint not null default 300` |
| `BidModel.maxAmount` | `max_amount bigint null`（出价人授权上限，代理加价模式用） |

`FindByID` 必须还原 `tradingMode` 与各项价格，否则策略无法工作。

### 7.4 异步消费者接入（`BidProcessor`）

`ProcessBid` 的 happy path 基本不变（`auction.PlaceBid` 已内部委托策略）。需补充 **close-on-accept** 路径：

```go
if placeErr := auction.PlaceBid(money); placeErr != nil {
    return placeErr
}
s := p.strategyResolver.ForMode(auction.TradingMode())

// 代理加价模式：解析并批量写入收敛后的代理出价（在 UoW 内）
if pb, ok := s.(interface {
    ResolveProxyBids(*model.AuctionModel, []model.BidModel, uint64, model.MoneyModel) ([]strategy.ProxyAction, error)
}); ok {
    existing, _ := uow.BidRepository().FindByAuction(ctx, auction.ID())
    actions, resolveErr := pb.ResolveProxyBids(auction, existing, cmd.UserID, money)
    if resolveErr != nil {
        return resolveErr
    }
    for _, act := range actions {
        proxyBid, _ := model.NewBidModel(auction.ID(), act.UserID, act.Amount)
        if _, crerr := uow.BidRepository().Create(ctx, proxyBid, uuid.NewString()); crerr != nil {
            return crerr
        }
    }
}

// 荷兰式/一口价：接受即结束
if s.ShouldCloseOnAccept() {
    bids, _ := uow.BidRepository().FindByAuction(ctx, auction.ID())
    if closeErr := auction.Close(bids); closeErr != nil {
        return closeErr
    }
    // dispatch AuctionEndedEvent（含 winner）
}

// 防狙击：每次成功出价后顺延结束时间（与交易模式正交）
auction.MaybeExtendEndTime(time.Now().UTC())

persistedBid, err := uow.BidRepository().Create(ctx, bid, cmd.IdempotencyKey)
// ... 其余不变（uow.Complete 提交；dispatch BidPlacedEvent）
```

为此给 `BidProcessor` 注入 `strategy.Resolver`（在 `module.go` 的 `fx.Provide(messaging.NewBidProcessor)` 处自动满足）。

`IsPermanentBidError` 需扩展，把新模式下的“永久业务错误”纳入 DLQ（如 `ErrDutchBidMustMatchPrice`、`ErrFixedPriceMismatch`、保留价相关错误），瞬时错误照旧延迟重试。

### 7.5 WebSocket 广播

- `AuctionEndedEvent` 已通过 WS Hub 广播；补充 `WinnerUserID` / `WinningAmount` 字段即可。
- 荷兰式实时降价：可选新增 `PriceDropEvent`，由定时任务（或读路径计算）推送当前价，前端展示倒计时降价。

---

## 8. 数据库迁移

在 `migrations/` 新增一个 `.sql`：

```sql
-- auctions 表：交易模式 + 价格参数 + 防狙击开关
ALTER TABLE auctions
    ADD COLUMN trading_mode       varchar not null default 'english',
    ADD COLUMN starting_price     bigint,
    ADD COLUMN price_step         bigint,
    ADD COLUMN reserve_price      bigint,
    ADD COLUMN current_price      bigint,
    ADD COLUMN winner_user_id     bigint,
    ADD COLUMN winning_bid_amount bigint,
    ADD COLUMN anti_snipe_enabled boolean not null default false,
    ADD COLUMN extension_window_sec bigint not null default 300,
    ADD CONSTRAINT chk_trading_mode CHECK (
        trading_mode IN ('english','dutch','sealed_bid','vickrey','fixed_price','ebay_proxy')
    );

-- bids 表：代理加价模式下每条出价携带出价人授权上限
ALTER TABLE bids
    ADD COLUMN max_amount bigint;
```

---

## 9. 如何扩展一种新的交易模式

开闭原则落地步骤（不改既有流程）：

1. 在 `domain/enum/trading_mode_enum.go` 增加常量 + 校验分支。
2. 在 `domain/strategy/` 新建 `xxx_strategy.go`，实现 `TradingStrategy` 全部方法。
3. 在 `module.go` 用 `fx.ResultTags(`group:"trading_strategy"`)` 注册。
4. 如需新列，`migrations/` 加迁移；`AuctionModel` 增加字段与 getter。
5. 为新错误在 `domain/errs/errs.go` 定义，并视情况加入 `IsPermanentBidError`（如 `ErrProxyMaxTooLow`）。
6. 写策略单测（见第 10 节）。
7. 代理加价（`ebay_proxy`）：在 `module.go` 注册 `EbayProxyAuctionStrategy`，并确认 `bids` 表有 `max_amount` 列、`BidModel` 有 `MaxAmount()`。
8. 防狙击（正交）：创建拍卖时支持 `antiSnipeEnabled` / `extensionWindowSec` 入参；`BidProcessor` 在每次出价成功后调用 `MaybeExtendEndTime`，无需注册为新策略。

---

## 10. 测试策略

复用仓库已有的 `*_test.go` 表驱动风格（`auction_model_test.go` 等）：

- **策略单测**：每个策略覆盖 `ValidateBid`（边界：首单/重复/低于当前价/保留价）、`DetermineWinner`（最高/第二价/流拍）、`ShouldCloseOnAccept`。
- **Resolver 单测**：未知模式返回 `ErrUnsupportedTradingMode`；重复注册应覆盖或报错。
- **模型单测**：`PlaceBid` 在 `english` 下行为不变（回归）；`Close(bids)` 正确写入 winner。
- **集成测试**：扩展 `bid_processor_test.go`，验证荷兰式出价后拍卖自动关闭且 winner 正确；密封/维克里在关闭时计算正确成交价。
- **代理加价单测**：`ResolveProxyBids` 覆盖——单出价人直接领先、双出价人反制循环收敛、max 不足出局、多人竞价最终次高出价者胜且付第二价。
- **防狙击单测**：`MaybeExtendEndTime` 在窗口内出价顺延、窗口外不顺延、未启用开关不顺延；`bid_processor_test` 验证窗口内出价后 `endTime` 被推后。

---

## 11. 风险与权衡

| 风险 / 权衡 | 说明 | 缓解 |
|---|---|---|
| 策略读取模型状态 | 策略通过 getter 读未导出字段，无法直接写；写入统一在模型 | 已通过 `PlaceBid`/`Close`/`ApplyWin` 收敛 |
| 荷兰式实时降价 | 按“已过去时间段”计算价格，依赖节点时钟 | 单实例/同区域部署可接受；跨时区需统一时钟源 |
| 向后兼容 | 存量拍卖无 `trading_mode` | 迁移默认 `english`，行为零变化 |
| DLQ 分类 | 新模式错误需正确归类为永久/瞬时 | 在 `IsPermanentBidError` 显式登记 |
| `Resolver()` 包级单例 | 策略解析走全局单例，测试可替换 | 提供 `SetResolverForTest` 钩子便于单测 |
| 代理加价收敛性 | `ResolveProxyBids` 反制循环在极端并发下需保证收敛 | 算法保证公开价严格递增且受 max 上界约束；整段在 UoW 事务 + `SELECT FOR UPDATE` 内执行，避免竞态 |
| 代理加价隐藏 max | 广播/查询不应泄露他人 max | 对外 DTO 与 `BidPlacedEvent` 仅暴露 `amount`；`maxAmount` 仅持久化与关闭判定使用 |
| 防狙击时钟 | `MaybeExtendEndTime` 依赖节点时钟，多实例需一致 | 同区域部署可接受；跨节点用统一时钟源（NTP/TrueTime） |

---

## 12. 实施清单（分阶段）

- **Phase 1（零行为变更）**：`TradingModeEnum` + `TradingStrategy` 接口 + `EnglishAuctionStrategy`（平移现有逻辑）+ `Resolver` + Fx 装配 + `AuctionModel` 加字段并委托策略。确保全部既有测试通过。
- **Phase 2（持久化）**：迁移 SQL + repository 映射 + `CreateAuctionCommand` 接入模式与价格参数。
- **Phase 3（close-on-accept）**：荷兰式 + 一口价，打通 `BidProcessor` 的 `ShouldCloseOnAccept` 路径与 `AuctionEndedEvent` 富化。
- **Phase 4（结束判定）**：密封投标 + 维克里，完善 `Close`/`DetermineWinner` 与 `CloseAuctionCommand`。
- **Phase 5（代理加价）**：`ebay_proxy` 策略 + `BidModel.maxAmount` + `bids.max_amount` 迁移 + `BidProcessor` 的 `ResolveProxyBids` 批量写入路径 + Fx 注册。
- **Phase 6（防狙击）**：`auctions` 表 `anti_snipe_enabled`/`extension_window_sec` + `CreateAuctionCommand` 入参 + `MaybeExtendEndTime` 在 `BidProcessor` 每次成功后调用。
- **Phase 7（体验与质量）**：荷兰式 `PriceDropEvent` 实时广播；代理加价前端 max 输入 + 防狙击倒计时提示；各策略单测/集成测试；更新 README 与本文档。

---

## 附：关键文件索引

| 文件 | 角色 |
|---|---|
| `internal/modules/auction/domain/model/auction_model.go` | 领域模型，`PlaceBid`/`Close`/`MaybeExtendEndTime` 改造点 |
| `internal/modules/auction/domain/model/bid_model.go` | 新增 `maxAmount` 支持代理加价模式 |
| `internal/modules/auction/domain/enum/auction_state_enum.go` | 枚举写法参考 |
| `internal/modules/auction/domain/enum/trading_mode_enum.go` | **新增** 交易模式枚举 |
| `internal/modules/auction/domain/strategy/` | **新增** 策略接口与各实现 + Resolver |
| `internal/modules/auction/infra/messaging/bid_processor.go` | 异步消费者，接入 close-on-accept |
| `internal/modules/auction/infra/messaging/bid_processor.go:IsPermanentBidError` | DLQ 错误分类扩展 |
| `internal/modules/auction/module.go` | Fx 装配策略 |
| `internal/modules/auction/application/command/create_auction_command.go` | 创建拍卖接入模式 |
| `migrations/` | 新增交易模式相关列的迁移 |
