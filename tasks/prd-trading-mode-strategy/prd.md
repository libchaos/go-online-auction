# PRD / Tech Spec — 基于策略模式的多交易模式实现

> 本目录的设计文档（合并 PRD + 技术规格）位于：`../../docs/trading-mode-strategy-design.md`
>
> 任务跟踪见同目录 `tasks.md`。

## 范围

在现有六边形架构（Ports & Adapters + CQRS + DDD + Uber Fx）拍卖系统上，引入**策略模式**
以支持多种交易模式，并在运行时按 `TradingModeEnum` 解析对应策略：

- `english` 英式拍卖（现行逻辑平移）
- `dutch` 荷兰式降价
- `sealed_bid` 密封投标（盲拍）
- `vickrey` 维克里（第二价密封）
- `fixed_price` 一口价
- `ebay_proxy` 类 eBay 自动代理加价

防狙击 / 延时关闭作为与交易模式**正交**的关注点，通过 `MaybeExtendEndTime` 接入。

## 关键约束（遵循 `.agent` 规范）

- 注释仅允许出现在接口声明上，函数体/实现不加注释。
- 命名不使用缩写（如 `ProxyResolvable`、`DetermineWinner`）。
- 枚举采用 `Enum{Name}{Value}` 模式 + `validate{Name}Enum` 映射 + `errors.New`。
- 领域错误统一为 `errs.ErrXxx`。
- 值对象不可变、按值返回。
- 测试使用 `testify` 的 `suite`，含 AAA 注释，单行 ≤ 120 字符。
