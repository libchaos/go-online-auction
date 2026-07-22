# 项目长期记忆：go-online-auction

## 环境（每次跑 Go 命令必看）
- Go 编译器**仅**在 Encore 捆绑路径：`/home/chaos/.encore/encore-go/bin/go`（go1.26.3-encore），位于 `Ubuntu-22.04` WSL 发行版。
- 所有 Go 命令必须经：`wsl -d Ubuntu-22.04 -e bash -c "export PATH=\$PATH:/home/chaos/.encore/encore-go/bin:/home/chaos/go/bin; export GOFLAGS=-mod=mod; cd /home/chaos/go-pro/go-online-auction && <cmd>"`。Windows Git Bash 原生 `go`/裸 `ls`/`grep` 无法解析 `\\wsl.localhost` 路径，必须在 wsl 内执行。
- 静态工具在 `/home/chaos/go/bin`：`golangci-lint`、`govulncheck`、`nilaway`。

## 项目概况
- Go + React 实时在线拍卖。六边形 + CQRS + 事件驱动 + DDD + Uber Fx 依赖注入，Cobra CLI。消息中间件 **NATS JetStream**（无 Redis）。HTTP 端口 **9000**（README 正文 8080 笔误）。
- 多交易模式（english/dutch/sealed_bid/vickrey/fixed_price/ebay_proxy）经策略模式 + `TradingModeEnum` 运行时解析；防狙击是与模式正交的关闭策略。
- 模块：`users`(auth+RBAC) / `auction` / `listing` / `deposit`+`ledger`(保证金/资金账本) / `payment`(当面付+支付宝提现 Saga) / `notification`(站内+SSE+邮件) / `watchlist`(SPU收藏)。
- 迁移号段：deposit `000009/10`、ledger `000012–15`、payment `000018–21`、notification `000022/23/24`、watchlist `000025`、RBAC `000026/27/28`。

## RBAC / Casbin（完整 g 表，已实施）
- 共享包 `internal/shared/modules/authz`：`model.conf` 内嵌 + `PgxAdapter` 落 `casbin_rules` + `RequirePermission()` 中间件 + `authz.Module`。matcher=`g(r.sub,p.sub) && keyMatch4(r.obj,p.obj) && (r.act==p.act||p.act=="*")`；`r.sub`=`claims.UserID`(strconv)。
- `keyMatch4` 陷阱：`/*`→多段、路径参数用 `{id}`（非 `:id`）、`**` 生成非法正则会 panic，**禁用**。
- 策略 seed `000026/27`（admin=`/api/v1/* *`；seller=拍卖/SPU/SKU 管理；bidder=保证金）。g 表 seed `000028` 从 `users.role` 回填 → **上线前必 `goose up` 应用 000026/27/28**，否则非 admin 全 403。
- 角色动态分配 API（admin 专属，`RequireRole(admin)` 保护）：`/api/v1/rbac/role-assignments`（POST 分配 / GET / DELETE 单个 / DELETE /{userID} 全部）。`users.role` 仅用于管理权与 JWT，资源鉴权走 g 表。
- 受保护路由用 `authz.RequirePermission()`（users/auction/deposit/listing）。logger 用项目 `logger.Logger`（zerolog）。

## 模块要点（精简）
- **deposit/ledger**：DB 为意图账本/状态机。`DepositUnitOfWork` 复用同一 `pgx.Tx`，5 原子操作转 ledger，`PaymentPort` 降为可选外部适配层。HTTP `/api/v1/ledger/*`。
- **payment**：充值当面付（回调→MQ→ledger 入账）+ 提现转支付宝 Saga。已发布 `payment.evt.deposit.success`、`payment.evt.withdrawal.completed/.failed`（扁平 payload）。stream `PAYMENT_EVENTS`。
- **notification**：纯消费者，SSE 实时推送（`?token=`，因 EventSource 不能带头）+ 通知中心 + 偏好（payment/deposit/auction/listing/system）。`SourceEventConsumer` 订阅 payment/deposit/auction/listing；邮件经 `notification_outbox`→`notification.evt.email.requested`→`EmailDispatchConsumer`。域枚举 `String()` 为**指针接收者**，调用前须先赋值本地变量。
- **watchlist**：迁移 `000025`，`/api/v1/watchlists`（SPU 级收藏，RequireAuth）。

## 规范（`.agent`）
- 注释**仅**允许在接口声明上；命名不用缩写；枚举 `Enum{Name}{Value}`+`validate{Name}Enum`+`errors.New`；领域错误 `errs.ErrXxx`；值对象不可变按值返回。
- 测试用 `testify` `suite` + AAA 注释，单行 ≤120 字符；mock 可选调用加 `.Maybe()`。
- `make static`=lint+govulncheck+nilaway。`govulncheck` 因**预先存在**依赖 CVE（pgx/chi/x-crypto）失败，与特性无关；特性门禁以 lint/nilaway/test 为准。`nilaway` 已 `--exclude-pkgs="vendor/,auction/tests/mocks"`（mockery 生成 UoW/接口 mock 故意返回可能 nil 值，会被误报；确认是 mock 来源后加包路径到 exclude，勿改生产代码）。

## Fx 装配约定（重要，易漏导致启动断链）
- **sqlcgen.DBTX**：模块若其 sqlcgen repository 直接吃 `sqlcgen.DBTX`，`Module` 必须 `fx.Provide(func(pool *pgxpool.Pool) sqlcgen.DBTX { return pool })`（users/listing/auction/notification 已提供；deposit/payment 曾漏，2026-07-22 补）。`go build` 查不出，`go run ./main.go all` 组装整图时才报 `missing type: sqlcgen.DBTX`。ledger 的 UoW factory 直接吃 `*pgxpool.Pool`，无需此 provider。
- **websocket.EventConsumer**（2026-07-22 又遇）：websocket `Hub` 依赖本地 `websocket.EventConsumer` 接口（签名 `Consume(ctx, handler func(subject,data []byte)) error`）。其实现（如 `messaging.NewJetStreamXxxEventConsumer`）必须 `fx.As(new(websocket.EventConsumer))` 显式绑定，否则 `NewHub` 报 `missing type: websocket.EventConsumer`。同模式参考：auction=`fx.As(new(messaging.EventConsumer))`、notification=`fx.As(new(sse.EventConsumer))`。
- **config 子配置抽取**（2026-07-22 又遇）：fx 图**只**提供根 `config.Config`（各 squash 子结构 `Email`/`Alipay`/`DB`/`Outbox`/`NATS`/`scheduler.Config`/各 `outbox.Config` 等）。凡构造器参数直接吃 `config.X` 子类型（如 `NewSMTPEmailAdapter(cfg config.Email)`、`NewPaymentPort(cfg config.Alipay)`），其所在模块**必须** `fx.Provide(func(cfg config.Config) X { return cfg.X })` 显式抽取，否则按 `missing type: config.X` 断链。notification 曾漏抽 `config.Email`、payment 已抽 `config.Alipay` 为同模式参考。

## 如何正确跑起来（2026-07-22 核实）
- **配置键名坑**：真实键名是 `DB_*`（见 `config/db.go` mapstructure），非 README 写的 `DATABASE_*`。`DB_SSL_MODE=false`→`sslmode=disable`（勿写 `disable` 字面量）。以 `.env.example` 为准。
- **docker-compose 已停用**（2026-07-22 注释）：改用本地 PostgreSQL + NATS（与 Go 同机，WSL 内）。`.env` 默认指向 `localhost`。
- 命令：`go run ./main.go all`（:9000）；子命令 `auction`/`websocket`/`db:migrate`(+`:down`/`:status`)/`create-admin --name --email --password`。
- **启动顺序**：①起 Postgres+NATS(-js) → ②`cp .env.example .env`（已存在则跳过）→ ③`db:migrate`（含 000028）→ ④`create-admin` → ⑤**g-RBAC 引导**：用 admin 调 `POST /api/v1/rbac/role-assignments {user_id,role:"admin"}` 给自己绑 g（否则资源端点 403）。
- NATS stream 由 `nats.Module.OnStart` 的 `CreateOrUpdateStreams` 自举，无需手动建。
- 前端（可选）：`cd frontend-demo && npm install && npm run dev -- --host`（:5173，`VITE_API_BASE_URL=http://localhost:9000/api/v1`）。
