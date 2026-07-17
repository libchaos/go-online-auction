# Implementation Task Summary for 基于策略模式的多交易模式实现

> 参考：本目录 `prd.md`（指向 `../../docs/trading-mode-strategy-design.md`）。使用 X.0 表示主任务，X.Y 表示子任务。
> 设计文档 §12 将实施分为 7 个阶段；本表据此跟踪完成度。

## Phases

- **Phase 1 — 零行为变更**：1.0
- **Phase 2 — 持久化**：2.0
- **Phase 3 — close-on-accept**：3.0
- **Phase 4 — 结束判定**：4.0
- **Phase 5 — 代理加价**：5.0
- **Phase 6 — 防狙击**：6.0
- **Phase 7 — 体验与质量**：7.0（后端质量，已完成）/ 7.1–7.4（体验与文档，待办）

## Dependency Map

| 任务 | 依赖 | 说明 |
|---|---|---|
| 1.0 枚举 + 接口 + 英式策略 + Resolver + Fx + 模型委托 | — | 不改变既有行为 |
| 2.0 迁移 + repository 映射 + CreateAuctionCommand 接入 | 1.0 | |
| 3.0 荷兰式/一口价 close-on-accept + AuctionEndedEvent 富化 | 1.0, 2.0 | |
| 4.0 密封/维克里 Close/DetermineWinner + CloseAuctionCommand | 1.0, 2.0 | |
| 5.0 ebay_proxy + maxAmount + ResolveProxyBids + Fx | 1.0, 2.0, 3.0 | |
| 6.0 防狙击列 + CreateAuctionCommand 入参 + MaybeExtendEndTime | 1.0, 2.0 | |
| 7.0 各策略单测（strategy_test.go） | 1.0–6.0 | 已完成 |
| 7.1 荷兰式 PriceDropEvent 实时广播 | 3.0 | 待办 |
| 7.2 前端：代理 max 输入 + 防狙击倒计时提示 | 5.0, 6.0 | 待办（frontend-demo） |
| 7.3 README 补充多交易模式说明 | 1.0–6.0 | 待办 |
| 7.4 交易模式专属集成测试 | 1.0–6.0 | 待办 |

## Tasks

* [x] 1.0 Phase 1（零行为变更）：`TradingModeEnum` + `TradingStrategy` 接口 + `EnglishAuctionStrategy`（平移现有逻辑）+ `Resolver`(`strategy.GetResolver()`) + Fx `fx.Group:"trading_strategy"` 装配 + `AuctionModel` 增加 `tradingMode` 字段并委托策略（`PlaceBid`/`Close`/`DetermineWinner`）
* [x] 2.0 Phase 2（持久化）：迁移 `000004_add_trading_mode_columns.up.sql` + `AuctionEntity`/`BidEntity` 映射 + `CreateAuctionCommand` 接入 `TradingMode` 与价格参数（含校验）
* [x] 3.0 Phase 3（close-on-accept）：`DutchAuctionStrategy` + `FixedPriceAuctionStrategy`，打通 `BidProcessor` 的 `ShouldCloseOnAccept` 路径并富化 `AuctionEndedEvent`
* [x] 4.0 Phase 4（结束判定）：`SealedBidAuctionStrategy` + `VickreyAuctionStrategy`（第二价），完善 `Close`/`DetermineWinner` 与 `CloseAuctionCommand`（按 `auction_id` 拉取出价后委托策略）
* [x] 5.0 Phase 5（代理加价）：`EbayProxyAuctionStrategy` + `BidModel`/`BidEntity` 的 `maxAmount` + `bids.max_amount_in_cents` 迁移 + `BidProcessor` 的 `ResolveProxyBids` 批量写入路径 + Fx 注册 + `IsPermanentBidError` 错误分类扩展
* [x] 6.0 Phase 6（防狙击）：`auctions` 表 `anti_snipe_enabled`/`extension_window_sec` 迁移列 + `CreateAuctionCommand` 入参 + `MaybeExtendEndTime` 在 `BidProcessor` 每次成功后调用（正交关注点）
* [x] 7.0 Phase 7 后端质量：各策略单测 `internal/modules/auction/domain/strategy/strategy_test.go`（覆盖 Mode/ValidateBid/SuggestNextPrice/DetermineWinner/ResolveProxyBids/ShouldCloseOnAccept/Resolver/Money），并更新 model/mapper/command/bid_processor 既有测试以适配新签名
* [ ] 7.1 Phase 7 体验：荷兰式 `PriceDropEvent` 实时广播（当前 `DecrementDutchPrice` 仅更新模型，未发事件）
* [ ] 7.2 Phase 7 体验：代理加价前端 max 输入 + 防狙击倒计时提示（`frontend-demo` 尚未接入交易模式字段）
* [ ] 7.3 Phase 7 文档：README 补充多交易模式 / 代理加价 / 防狙击说明
* [ ] 7.4 Phase 7 测试：交易模式专属集成测试（现有 `tests/integration` 通过，但无针对各模式的专项用例）

## 质量门禁（完成 1.0–7.0 时）

| 门禁 | 结果 |
|---|---|
| `go build ./...` | ✅ 通过 |
| `golangci-lint run ./...` | ✅ 0 issues（含修复一处既有 nestif：`bid_repository.go` 提取 `mapPostgresCreateError`） |
| `nilaway` | ✅ 0 issues |
| `go test ./...` | ✅ 全部包 + 集成测试通过 |
| `govulncheck` | ⚠️ 仅因**预先存在**的依赖 CVE 失败（pgx v5.7.4 / chi v5.2.3 / golang.org/x/crypto openpgp），与本次特性无关，建议单独依赖升级 PR 处理 |
