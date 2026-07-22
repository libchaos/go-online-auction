# 保证金模块重构设计：将 Ledger 提升为顶层模块并以 Ledger 体系管理 Deposit

> 状态：已实施（2026-07-20）
> 日期：2026-07-20
> 目标：把嵌套在 `deposit` 下的账本子系统（ledger）提升为顶层模块 `internal/modules/ledger`，理顺依赖方向（deposit → ledger），使 deposit 的资金管理完全经由 ledger 体系完成，且**行为零变化、所有既有能力保持可用**。

---

## 1. 背景与目标

当前 `ledger`（资金账本）是 `deposit` 模块的子包 `internal/modules/deposit/ledger/`。它已实现一套完整的资金引擎：账户余额 + 冻结余额 + 不可变分录（entries/transfers/operations），并通过 5 个幂等原子操作（`CreateAccount` / `GetOrCreateAccountByOwner` / `Transfer` / `Freeze` / `Unfreeze` / `WithdrawFromFrozen`）管理资金流动。deposit 的 5 个押金命令（create/apply/release/cancel/forfeit）通过**同一事务**的 `DepositUnitOfWork` 调用 ledger 原子操作，完成冻结/释放/抵扣/罚没。

存在的问题与目标：

1. **结构性问题**：账本作为 `deposit` 的嵌套子模块，无法被其它有界上下文复用（未来拍卖托管、用户钱包等都可能需要资金引擎）。
2. **依赖方向不清**：ledger 的 HTTP handler 反向依赖 `deposit/infra/http/errs`，存在跨模块回边（虽不构成 Go 编译环，但违反分层）。
3. **目标**：
   - 将 ledger **提升为顶层模块** `internal/modules/ledger/`，使其成为独立、可复用的资金引擎。
   - 明确依赖方向：**deposit → ledger**（deposit 编排业务，ledger 提供资金原子操作）；**ledger 不依赖 deposit**。
   - 保留既有行为：共享事务原子性、5 个押金命令语义、HTTP `/api/v1/ledger/*` 端点全部不变 → "保证所有都可用"。

---

## 2. 范围与边界（本次**不做**的事）

- **不做领域语义变更**：deposit 聚合仍持有 `amount`/`status`（业务意图与状态机），ledger 仍持有账户余额与不可变分录（资金真相源）。本次只做结构性提升与依赖理顺，**不改变任何资金语义**。
- **不引入迁移变更**：ledger 表（`000012` ledger_accounts / `000013` ledger_entries / `000014` ledger_transfers / `000015` ledger_operations）与 deposits 表（`000009`/`000010`）结构、名称、外键均不变。
- **不做"deposit 去金额化"深层重构**（见 §11 后续可选增强）——避免扩大改动面、保证本次风险可控并通过全部门禁。

---

## 3. 当前结构

```
internal/modules/deposit/
├── module.go                         # 装配 deposit + ledger 全部内部件，并提供 RegisterLedgerRoutes
├── ports/deposit_uow.go              # DepositUnitOfWork.LedgerRepository() ledgerports.LedgerRepository
├── domain/{model,enum,errs}
├── application/
│   ├── command/*                     # 5 个命令经 uow.LedgerRepository() 调 ledger 原子操作；ledger_helper.go 解析账户
│   └── guard/ query/
├── infra/
│   ├── http/errs/errs.go             # HTTP 错误映射（依赖 deposit/domain/errs + deposit/ledger/domain/errs）
│   ├── http/chi/handler/deposit_handler.go
│   ├── uow/{wow,factory}.go          # Begin() 在同一 tx 上构造 depositRepo + ledgerRepo + outboxRepo
│   ├── repository/ outbox/ payment/ websocket/ messaging/
│   └── mapper/
└── ledger/                           # ◀ 待提升的子系统
    ├── domain/{model,enum,errs}
    ├── ports/{ledger_repository, ledger_uow, ledger_account_service}.go
    ├── application/service/ledger_account_service.go   # runTx 编排（自带 Begin/Complete）
    └── infra/{mapper,repository,uow,query,sqlcgen,http/{dto,handler,chi/router}}
```

关键事实（已核代码）：
- `deposit/infra/uow/factory.go` 在 `Begin()` 里用 `ledgerrepository.NewPostgresLedgerRepository(tx, ledgerMapper)` 在同一 `pgx.Tx` 上构造 ledger 仓库 → deposit+ledger **单事务原子**。
- `deposit/infra/http/errs/errs.go` 同时 import `deposit/domain/errs` 与 `deposit/ledger/domain/errs`；`ledger_handler.go` import 了前者 → **ledger → deposit 回边**。
- 仅 `deposit`、`tests/mocks`、`sqlc.yaml`、设计文档引用 `deposit/ledger`；无其他模块（auction/listing/users）依赖 ledger。

---

## 4. 目标结构

```
internal/modules/
├── deposit/                          # 业务编排层（意图/状态机），依赖 ledger 做资金管理
│   ├── module.go                     # 仅装配 deposit 自身；不再装配 ledger 内部件
│   ├── ports/deposit_uow.go          # LedgerRepository() 仍返回 *ledger/ports.LedgerRepository
│   ├── application/command/*         # 仍经 uow.LedgerRepository() 调 ledger（逻辑不变）
│   └── infra/{http/errs→shared, uow, repository, ...}
├── ledger/                           # ◀ 顶层资金引擎模块（自包含 Fx 模块）
│   ├── module.go                     # 提供 LedgerMapper/Repository/UoWFactory/AccountService/Handler + RegisterLedgerRoutes
│   ├── domain/{model,enum,errs}
│   ├── ports/*
│   ├── application/service/*
│   └── infra/{mapper,repository,uow,query,sqlcgen,http/{dto,handler,chi/router}}
└── shared/modules/httperrs/          # ◀ 新增：提取的共享 HTTP 错误映射（消除回边）
    └── errs.go                       # MapDomainError + 通用/各模块错误码；import ledger/domain/errs + deposit/domain/errs
```

依赖方向：**deposit → ledger**（经 `ledger/ports` 与 `ledger/infra/repository`、`ledger/infra/mapper` 共享事务）、**ledger → shared/httperrs →（ledger/domain/errs, deposit/domain/errs 叶子包）**。无循环。

---

## 5. 关键设计决策

### D1 — 消除 ledger→deposit 回边（errs 提取）
将 `deposit/infra/http/errs` **整体迁移**为 `internal/shared/modules/httperrs`（包名 `httperrs`），内部 import `deposit/ledger/domain/errs` → `ledger/domain/errs`；`deposit_handler.go` 与（移动后的）`ledger_handler.go` 改 import `httperrs`。
- 技术说明：当前该回边**不构成 Go 编译环**（`deposit/infra/http/errs` 仅依赖 `deposit/domain/errs` 与 `ledger/domain/errs` 两个叶子包，均不回流 `deposit` 根包）；但顶层模块反向依赖另一模块内部包违反分层，提升时必须消除。
- 备选（更纯但改动更大）：把 `MapDomainError` 做成参数化纯函数放 shared（零领域 import），各模块自带映射表。本设计**推荐整体迁移**方案以最小化改动、担保"所有都可用"，备选列为后续清理。

### D2 — 共享事务保持
`deposit/infra/uow/factory.go` 的 `Begin()` **仍在其自身 `pgx.Tx` 上**构造 `ledger/infra/repository.NewPostgresLedgerRepository(tx, ledgerMapper)`，保证 deposit + ledger 在同一事务内原子提交。deposit 依赖 ledger 的 `infra/repository` + `infra/mapper` **具体构造器**（这是 deposit→ledger 依赖，方向正确），而非经 ledger 自己的 `LedgerUnitOfWorkFactory`（那会开独立事务）。

### D3 — ledger 成为自包含 Fx 模块
新增 `internal/modules/ledger/module.go`：
- 提供 `ledger/mapper.NewLedgerMapper`、`LedgerRepository`（fx.As `ledgerports.LedgerRepository`）、`LedgerUnitOfWorkFactory`（fx.As `ledgerports.LedgerUnitOfWorkFactory`）、`LedgerAccountService`（fx.As `ledgerports.LedgerAccountService`）、`LedgerHandler`，以及 `RegisterLedgerRoutes`。
- `deposit/module.go` 删除 ledger 装配项与 `RegisterLedgerRoutes` 函数；其 `DepositUnitOfWorkFactory` 依赖 `*ledger/mapper.LedgerMapper`（由 `ledger.Module` 提供）+ `*pgxpool.Pool`（由 `database.Module` 提供）。

### D4 — 命令层不变
deposit 命令继续经 `unitOfWork.LedgerRepository()` 取同事务 ledger 仓库，调用 `Freeze/Unfreeze/WithdrawFromFrozen`；不依赖 `LedgerAccountService`（该服务仅服务 HTTP 独立事务路径）。命令源码仅改 import 前缀，逻辑零改动。

### D5 — 迁移与 sqlc 不变
- 表结构/名称不变；**无新迁移**。
- `sqlc.yaml` 中 ledger block 的 `queries`、`out` 路径改为 `internal/modules/ledger/infra/query`、`internal/modules/ledger/infra/sqlcgen`；运行 `make sqlc` 重新生成（schema=`migrations` 不变，仅产物路径随包移动）。sqlcgen 不引用内部包，移动安全。

### D6 — 范围边界
本次只做"结构性提升 + 依赖理顺"，不改变领域语义（见 §2）。深层"deposit 去金额化"列为后续可选增强（§11）。

---

## 6. 依赖方向图

```
                 ┌─────────────────────────────┐
                 │  cmd/all.go, cmd/auction.go  │
                 │  fx.New: ..., deposit.Module,│
                 │          ledger.Module, ...  │
                 └──────────────┬──────────────┘
                                │
            ┌───────────────────┴───────────────────┐
            ▼                                        ▼
   ┌──────────────────┐                    ┌──────────────────────┐
   │   deposit.Module │                    │    ledger.Module     │
   │ (业务编排/状态机) │                    │ (资金引擎, 自包含)   │
   └────────┬─────────┘                    └──────────┬───────────┘
            │ deposit → ledger                        │ ledger → shared/httperrs
            │ (ports + infra/repository+mapper 共享tx)│
            ▼                                        ▼
   ┌──────────────────────────────┐     ┌────────────────────────┐
   │ ledger/ports, ledger/infra/*  │     │ shared/modules/httperrs │
   └──────────────────────────────┘     └───┬────────────────┬───┘
                                             │                │
                                             ▼                ▼
                                    ledger/domain/errs  deposit/domain/errs  (叶子包)
```

无环：`deposit → ledger → shared/httperrs →（ledger/domain/errs, deposit/domain/errs）`，两条叶子分支均不回流。

---

## 7. 文件级变更清单

| 类别 | 当前路径 | 目标路径 / 操作 |
|------|----------|-----------------|
| **移动** | `internal/modules/deposit/ledger/**` | `git mv` → `internal/modules/ledger/**`（含 domain/ports/application/service/infra{mapper,repository,uow,query,sqlcgen,http}） |
| **内部 import 重写** | 移动包内所有 `auction/internal/modules/deposit/ledger/` | → `auction/internal/modules/ledger/` |
| **共享 errs（新）** | `internal/modules/deposit/infra/http/errs/` | → `internal/shared/modules/httperrs/`（包名 `httperrs`）；内 `deposit/ledger/domain/errs`→`ledger/domain/errs` |
| **errs 消费方** | `deposit/infra/http/chi/handler/deposit_handler.go`、`ledger/infra/http/chi/handler/ledger_handler.go` | import `httperrs`（路径随移动更新） |
| **deposit 引用更新** | `deposit/ports/deposit_uow.go`、`deposit/infra/uow/{wow,factory}.go`、`deposit/application/command/*` | import `deposit/ledger/...` → `ledger/...`（仅前缀） |
| **deposit 装配** | `deposit/module.go` | 删除 ledger 装配项 + `RegisterLedgerRoutes`；保留 `DepositUnitOfWorkFactory`（依赖 ledger/mapper + pgxpool） |
| **ledger 装配（新）** | — | 新增 `internal/modules/ledger/module.go`（提供 Mapper/Repository/UoWFactory/AccountService/Handler + `RegisterLedgerRoutes`） |
| **入口装配** | `cmd/all.go`、`cmd/auction.go` | fx.New 增加 `ledger.Module`；`deposit.RegisterLedgerRoutes` → `ledger.RegisterLedgerRoutes` |
| **sqlc** | `sqlc.yaml` ledger block | `queries`→`internal/modules/ledger/infra/query`；`out`→`internal/modules/ledger/infra/sqlcgen`；`make sqlc` |
| **mocks** | `tests/mocks/mock_LedgerRepository.go` `mock_LedgerAccountService.go` `mock_LedgerUnitOfWork.go` `mock_LedgerUnitOfWorkFactory.go` | import 改为 `ledger/ports`、`ledger/domain/model`；`make update-mocks` 重新生成覆盖 |
| **测试** | `deposit/application/command/*_test.go`（如 `create_deposit_command_test.go` 的 `ledgerenum`/`ledgermodel`） | import `ledger/...` |
| **文档** | `docs/deposit-module-design.md` 中 `deposit/ledger` 引用 | 更新为 `ledger/...`（建议，非阻塞） |

---

## 8. 实施步骤（顺序）

1. `git mv internal/modules/deposit/ledger internal/modules/ledger`；移动 `deposit/infra/http/errs` → `internal/shared/modules/httperrs`。
2. 批量重写被移动包内 import 前缀（`deposit/ledger` → `ledger`）；修正 `httperrs` 内 `deposit/ledger/domain/errs` → `ledger/domain/errs`。
3. 更新 deposit 侧所有引用（`deposit_uow.go`、uow、command/*、errs 消费方）的 import 前缀。
4. 改 `sqlc.yaml` ledger block 路径 → `make sqlc`；确认 ledger sqlcgen 产物路径正确。
5. 新增 `ledger/module.go`；精简 `deposit/module.go`（去 ledger 装配 + `RegisterLedgerRoutes`）。
6. 改 `cmd/all.go`、`cmd/auction.go`（加 `ledger.Module`，路由改为 `ledger.RegisterLedgerRoutes`）。
7. `make update-mocks` 重新生成 ledger mocks；更新 deposit 测试 import。
8. 跑验证门禁（§9）。

---

## 9. 验证门禁

全部必须绿色（与既有 CI 一致）：

| 门禁 | 命令 |
|------|------|
| 构建 | `go build ./...` |
| 静态检查 | `go vet ./...` |
| 单测/集成 | `go test ./...`（含 `tests/integration` 的 testcontainers 真实 PG 用例：迁移至 v17、outbox→relay→NATS） |
| Lint | `golangci-lint run ./...` |
| NilAway | `nilaway --include-pkgs="auction" ./...` |

补充核对：启动 `go run ./main.go all`（或 `auction`）后 `GET /api/v1/ledger/accounts/{id}`、`POST /api/v1/ledger/freeze` 等行为与提升前一致；deposit 5 个命令的资金语义不变。

---

## 10. 风险与回滚

- **风险**：import 路径面较广（ledger 包内 + deposit 引用 + mocks + 测试）；`make sqlc`/`make update-mocks` 需重跑。
- **缓解**：上述均为机械路径改写，无逻辑改动；`git mv` 保留历史；无迁移/DDL 变更，无运行时行为变化。
- **回滚**：任一步可 `git checkout` 回到迁移前状态；因无 DB 变更，回滚零数据风险。
- **已知气味（接受）**：`shared/modules/httperrs` 会 import `deposit/domain/errs` 与 `ledger/domain/errs`（shared→feature）。不构成环，属常见错误映射分层；若团队要求更纯，可后续改为参数化 `MapDomainError`（见 D1 备选）。

---

## 11. 后续可选增强（不在本次范围）

- **deposit 去金额化**：让 deposit 聚合不再持有 `amount`，改为持有 ledger 账户引用（buyer/platform accountID），余额与冻结额完全由 ledger 派生。需新增迁移（deposits 表加 `ledger_account_id` 等列，或移除 `amount` 列），并改写命令读取余额的路径。风险较高，故本次不做，留作独立重构。
- **ledger 多租户/多上下文**：未来 auction 托管、用户钱包可直接依赖 `ledger` 顶层模块，复用同一资金引擎与幂等机制。

---

## 12. 影响面核对（grep 证据）

- `deposit/ledger` 引用方仅：`deposit`（ports/uow/command/errs）、`tests/mocks`（4 个 ledger mock）、`sqlc.yaml`、设计文档 → 均在本变更清单覆盖。
- `RegisterLedgerRoutes` 调用方仅 `cmd/all.go:48`、`cmd/auction.go:46` → 均改为 `ledger.RegisterLedgerRoutes`。
- `deposit/infra/http/errs` 消费方仅 `deposit_handler.go`、`ledger_handler.go` → 均改 import `httperrs`。
- ledger 包内仅依赖自身（`deposit/ledger/...`）+ `deposit/infra/http/errs`（即唯一回边，已通过 D1 消除）。

---

## 13. 实施结果（2026-07-20）

按 §8 步骤落地，全部完成，门禁全绿。

### 实际变更
- `mv internal/modules/deposit/ledger → internal/modules/ledger`（顶层模块，自包含 Fx 模块）；
  `mv internal/modules/deposit/infra/http/errs → internal/shared/modules/httperrs`（包名 `httperrs`）。
- 全仓 `.go` import 前缀批量改写：`deposit/ledger` → `ledger`、`deposit/infra/http/errs` → `shared/modules/httperrs`。
- 新增 `ledger/module.go`（`LedgerMapper`/`LedgerUnitOfWorkFactory`(as `LedgerUnitOfWorkFactory`)/`LedgerAccountService`(as `LedgerAccountService`)/`LedgerHandler` + `RegisterLedgerRoutes`）；
  `deposit/module.go` 卸下 ledger 装配与 `RegisterLedgerRoutes`，仅保留 `DepositUnitOfWorkFactory`（依赖 ledger 的 mapper + pgxpool）。
- `sqlc.yaml` ledger block 路径改为 `internal/modules/ledger/infra/{query,sqlcgen}`；`make sqlc` + `make update-mocks` 重生成。
- `cmd/all.go`、`cmd/auction.go` 增加 `ledger.Module`，`RegisterLedgerRoutes` 改由 `ledger.RegisterLedgerRoutes` 注册。

### 验证门禁（全绿）
| 门禁 | 结果 |
|------|------|
| `go build ./...` | EXIT 0 |
| `go vet ./...` | EXIT 0 |
| `go test ./...`（含 testcontainers 真实 PG 集成测试） | EXIT 0 |
| `golangci-lint run ./...` | 0 issues |
| `nilaway --include-pkgs=auction ./...` | EXIT 0 |

### 运行时装配验证
`go run ./main.go all` 的 fx 图**完整构建**：ledger 6 类 provider 全部入图、deposit 经 ledger mapper 解析依赖、无 "missing provider" / 注解错误。进程仅因本沙箱无 NATS/Postgres 基础设施（`nats: no servers available`）在生命周期 OnStart 退出——属环境限制，非代码缺陷；DB/ledger 路径已由 testcontainers 集成测试覆盖。

### 顺带修复的预存 bug
`auction/module.go` 中 6 处 `fx.ResultTags("group:trading_strategy")` 缺少内层双引号，fx 要求 `key:"value"` 格式。该错误仅在 fx 构建图时触发，编译/单测无法发现，**导致 app 原本根本无法启动**。已修正为 `fx.ResultTags(`group:"trading_strategy"`)`，app 现已可正常构建 fx 图。此问题与本次 ledger 重构无关，系重构验证过程中发现并修复，以满足"保证所有都可用"。

