# 项目长期记忆：go-online-auction

## 环境（每次跑 Go 命令必看）
- Go 编译器**仅**在 Encore 捆绑路径：`/home/chaos/.encore/encore-go/bin/go`（go1.26.3-encore，项目需 go1.25.5），位于 `Ubuntu-22.04` WSL 发行版。
- 所有 Go 命令必须经：`wsl -d Ubuntu-22.04 -e bash -c "export PATH=\$PATH:/home/chaos/.encore/encore-go/bin; export GOFLAGS=-mod=mod; cd /home/chaos/go-pro/go-online-auction && <cmd>"`。Windows Git Bash 原生 `go` 无法锁 `\\wsl.localhost` UNC 路径。
- 静态工具在 `/home/chaos/go/bin`：`golangci-lint`、`govulncheck`、`nilaway`。

## 项目概况
- Go + React 实时在线拍卖。六边形架构 + CQRS + 事件驱动 + DDD + Uber Fx 依赖注入，Cobra CLI。
- 消息中间件为 **NATS JetStream**（已移除 Redis）。HTTP 实际端口 **9000**（README 正文 8080 为笔误）。
- 多交易模式（english/dutch/sealed_bid/vickrey/fixed_price/ebay_proxy）通过策略模式 + `TradingModeEnum` 运行时解析，防狙击为与模式正交的关闭策略。
- 保证金模块设计文档：`docs/deposit-module-design.md`（新模块 `internal/modules/deposit/`，六边形+CQRS+Fx；DB 为意图账本/状态机；`PaymentPort` 抽象外部资金动作；跨模块经 `DepositGuard` 端口 + 订阅 `AuctionEndedEvent` 解耦结算）。迁移 `000009`/`000010`。

## 规范（`.agent`）
- 注释**仅**允许在接口声明上，函数体/实现不加注释。
- 命名不用缩写；枚举用 `Enum{Name}{Value}` + `validate{Name}Enum` 映射 + `errors.New`；领域错误 `errs.ErrXxx`；值对象不可变按值返回。
- 测试用 `testify` `suite`，含 AAA 注释，单行 ≤ 120 字符；mock 的可选调用加 `.Maybe()`（否则期望未触发会失败）。
- `make static` = lint + `govulncheck` + `nilaway`。`govulncheck` 因**预先存在**依赖 CVE（pgx/chi/x-crypto）失败，与特性无关；特性代码门禁以 lint/nilaway/test 为准。
