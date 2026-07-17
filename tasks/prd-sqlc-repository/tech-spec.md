# Technical Specification — 数据访问层重构：手写 pgx SQL 迁移到 sqlc

## Executive Summary

当前所有仓储（`auction`、`bid`、`user`、`refresh_token`）直接使用 pgx v5 手写 SQL：查询字符串内联在仓储方法中，`Scan` 按位置逐列映射到手写的 `infra/entity` 结构体。四个仓储合计约 900 行，其中 60% 以上是重复的列清单与 `Scan` 样板代码，且列顺序与 `Scan` 参数顺序的一致性完全靠人工维护（`bid_repository.go` 的 `FindByID` 已经漏扫了 `max_amount_in_cents`，属于此类隐患的实例）。

本方案引入 **sqlc**（`sqlc-dev/sqlc`，`sql_package: pgx/v5`）：SQL 移到独立的 `.sql` 查询文件，sqlc 在编译期根据 `migrations/` 中的 schema 做类型检查并生成类型安全的 Go 代码。**端口（`ports.*Repository`）、领域模型、UoW 语义、错误映射全部保持不变**——sqlc 只替换仓储内部的"SQL 执行 + Scan"这一层，生成的 `Queries` 天然同时支持 `*pgxpool.Pool` 与 `pgx.Tx`，与现有 UoW/事务模式无缝衔接。

收益：
- **编译期 SQL 校验**：`sqlc generate` 时对照 schema 检查列名、类型、参数个数；列增删后忘改 Scan 的整类 bug 消失。
- **消除样板**：手写的 entity 结构体与逐列 Scan 由生成代码替代，仓储只剩"调用查询 + 错误映射 + 领域转换"。
- **schema 单一事实源**：sqlc 直接解析 goose 格式的 `migrations/*.sql`（忽略 `-- +goose Down` 段），迁移即 schema，无需重复维护 DDL。

## 现状盘点

### 数据访问链路

```
command/query (application)
      │  依赖 ports.XxxRepository / ports.AuctionUnitOfWork(Factory)
      ▼
PostgresXxxRepository (infra/repository)     ← 手写 SQL + Scan（本次重构对象）
      │  依赖 uow.DBExecutor（*pgxpool.Pool 与 pgx.Tx 的公共接口）
      ▼
pgxpool.Pool / pgx.Tx
```

### 现有仓储与查询清单

| 仓储 | 方法 | 备注 |
|---|---|---|
| `PostgresAuctionRepository` | Create / FindByID / FindByIDForUpdate / Update / FindAllPaginated / Count | `FindByIDForUpdate` 用 `FOR UPDATE NOWAIT`，映射 `55P03` → `ErrConcurrencyConflict`；`FindAllPaginated`/`Count` 按 `state` 是否为 nil 走两条 SQL |
| `PostgresBidRepository` | Create / FindByID / FindByAuctionID / Update / FindTopBidsByAuctionID | Create 映射 `23514`（触发器）→ `ErrBidMustExceedHighest`、`23505`（幂等键）→ `ErrBidDuplicateIdempotencyKey` |
| `PostgresUserRepository` | Create / FindByID / FindByEmail / Update / FindAllPaginated / Count | `23505` → `ErrEmailAlreadyExists`；乐观锁 `WHERE id AND version` |
| `PostgresRefreshTokenRepository` | Create / FindByTokenHash / Update / RevokeAllForUser | — |

### 关键约束

1. **UoW 模式必须保留**：`AuctionUnitOfWorkFactory.Begin()` 开启 `pgx.Tx`，把同一个 tx 注入 auction/bid 两个仓储；`PlaceBid` 依赖 `FOR UPDATE NOWAIT` + 触发器 + 幂等键的多层并发控制，语义一字不能变。
2. **领域错误映射留在仓储层**：`pgx.ErrNoRows` / `pgconn.PgError`（23505、23514、55P03）到领域 `errs.*` 的翻译是仓储职责，sqlc 不涉足。
3. **entity 使用 `uint64` / `*uint64`**：Postgres `BIGINT` 在 sqlc 默认生成 `int64`，需要通过 overrides 或 mapper 层转换（见下文决策）。

## 设计方案

### 1. sqlc 配置与目录结构

```
sqlc.yaml                                    # 仓库根目录
internal/modules/auction/infra/
├── query/                                   # 手写 SQL 查询文件（sqlc 输入）
│   ├── auction.sql
│   └── bid.sql
├── sqlcgen/                                 # 生成代码（sqlc 输出，提交到 git）
│   ├── db.go                                # DBTX 接口 + Queries + WithTx
│   ├── models.go                            # 表结构体（替代手写 entity）
│   ├── auction.sql.go
│   └── bid.sql.go
├── repository/                              # 仓储：调用 sqlcgen + 错误映射（保留）
└── mapper/                                  # sqlcgen model ↔ domain model（改造）

internal/modules/users/infra/
├── query/
│   ├── user.sql
│   └── refresh_token.sql
├── sqlcgen/
└── ...（同上）
```

`sqlc.yaml`（单文件，两个生成单元，schema 共用 goose 迁移）：

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "migrations"          # sqlc 原生支持 goose 注释格式，自动忽略 Down 段
    queries: "internal/modules/auction/infra/query"
    gen:
      go:
        package: "sqlcgen"
        out: "internal/modules/auction/infra/sqlcgen"
        sql_package: "pgx/v5"
        emit_pointers_for_null_types: true   # 可空列生成 *T 而非 pgtype，贴近现有 entity
        overrides:
          - db_type: "pg_catalog.int8"
            go_type: "int64"
  - engine: "postgresql"
    schema: "migrations"
    queries: "internal/modules/users/infra/query"
    gen:
      go:
        package: "sqlcgen"
        out: "internal/modules/users/infra/sqlcgen"
        sql_package: "pgx/v5"
        emit_pointers_for_null_types: true
```

**决策：按模块拆分生成单元**。auction 与 users 各得一份独立的 `Queries`/`models`，模块间零共享生成代码，符合现有 DDD 模块边界（模块只通过 ports 交互，infra 互不可见）。代价是 `DBTX` 接口每个模块生成一份——这正好替代现有 `uow.DBExecutor` 的角色，且两者方法签名完全一致（`Exec/Query/QueryRow`）。

### 2. 类型映射决策

sqlc + `pgx/v5` + `emit_pointers_for_null_types` 的默认输出与现有 entity 的差异：

| 列类型 | 现有 entity | sqlc 生成 | 处理 |
|---|---|---|---|
| `BIGSERIAL` / `BIGINT NOT NULL` | `uint64` | `int64` | **mapper 层转换**（`uint64(m.ID)`）。不用 overrides 强转 uint64：金额/ID 语义上非负，但 DB 存的就是 int64，边界转换收敛到 mapper 一处 |
| `BIGINT`（可空） | `*uint64` | `*int64` | 同上，mapper 转换 |
| `TIMESTAMPTZ` | `time.Time` | `time.Time`（pgx/v5 原生） | 无需处理 |
| `auction_state` 等 enum | `string` | sqlc 生成 Go 枚举类型（`AuctionState`） | mapper 层 `string(m.State)`；生成的枚举常量顺便可替换散落的字符串字面量 |
| `varchar` 可空 | `*string` | `*string` | 无需处理 |

**决策：删除手写 `infra/entity` 包，mapper 直接消费 sqlc 生成的 model**。entity 与生成 model 字段一一对应，保留两层纯属重复。mapper 的 `ToDomain`/`ToEntity` 改为 `ToDomain(sqlcgen.Auction)` / 拆解 domain 到查询参数结构体（`sqlcgen.CreateAuctionParams`）。

### 3. 查询文件设计

每个现有 SQL 一条命名查询，`FindAllPaginated`/`Count` 的动态分支拆成两条静态查询（现有代码本来就是 if/else 两条 SQL，只是共享了方法）：

`internal/modules/auction/infra/query/auction.sql`（节选）：

```sql
-- name: CreateAuction :one
INSERT INTO auctions (listing_id, end_time, state, trading_mode, starting_price, price_step, reserve_price,
    current_price, highest_bid_amount_in_cents, winner_user_id, winning_bid_id, winning_bid_amount,
    anti_snipe_enabled, extension_window_sec, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: GetAuctionByID :one
SELECT * FROM auctions WHERE id = $1;

-- name: GetAuctionByIDForUpdate :one
SELECT * FROM auctions WHERE id = $1 FOR UPDATE NOWAIT;

-- name: UpdateAuction :execrows
UPDATE auctions
SET listing_id = $1, start_time = $2, ..., version = $16, updated_at = $17
WHERE id = $18 AND version = $19;

-- name: ListAuctions :many
SELECT * FROM auctions ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListAuctionsByState :many
SELECT * FROM auctions WHERE state = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountAuctions :one
SELECT COUNT(*) FROM auctions;

-- name: CountAuctionsByState :one
SELECT COUNT(*) FROM auctions WHERE state = $1;
```

要点：
- `Create*` 从 `RETURNING id` 改为 `RETURNING *`：直接拿到含 DB 默认值的完整持久化行，省去回填。
- `Update*` 用 `:execrows`：返回 `int64` 行数，乐观锁 `RowsAffected() == 0 → ErrConcurrencyConflict` 判断照旧。
- 触发器错误（23514）、幂等键冲突（23505）依然从 `error` 里 `errors.As` 提取 `pgconn.PgError`——sqlc 生成代码返回原始 pgx 错误，现有 `mapPostgresCreateError` / `isPgLockError` / `isUniqueViolation` 原样保留。
- 参数超过 3 个时 sqlc 自动生成 `XxxParams` 结构体，17 个位置参数的 `CreateAuction` 变成具名字段赋值，可读性显著提升。

### 4. 仓储改造模式

以 `FindByIDForUpdate` 为例，改造前 45 行，改造后：

```go
type PostgresAuctionRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.AuctionMapper
}

func NewPostgresAuctionRepository(db sqlcgen.DBTX, mapper *mapper.AuctionMapper) *PostgresAuctionRepository {
	return &PostgresAuctionRepository{q: sqlcgen.New(db), mapper: mapper}
}

func (r *PostgresAuctionRepository) FindByIDForUpdate(ctx context.Context, id uint64) (model.AuctionModel, error) {
	row, err := r.q.GetAuctionByIDForUpdate(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AuctionModel{}, errs.ErrAuctionNotFound
		}
		if isPgLockError(err) {
			return model.AuctionModel{}, errs.ErrConcurrencyConflict
		}
		return model.AuctionModel{}, err
	}
	return r.mapper.ToDomain(row)
}
```

构造函数参数从 `uow.DBExecutor` 换成 `sqlcgen.DBTX`（方法集相同，`*pgxpool.Pool` 与 `pgx.Tx` 均满足）。**`internal/shared/modules/uow/db_executor.go` 的 `DBExecutor` 接口随之删除**，`TxManager` 保留。

### 5. UoW 衔接

`AuctionUnitOfWorkFactory.Begin()` 改动仅一处——仓储构造入参从 `tx` 变为 `tx`（`pgx.Tx` 满足 `sqlcgen.DBTX`），其余不动：

```go
func (f *AuctionUnitOfWorkFactory) Begin(ctx context.Context) (ports.AuctionUnitOfWork, error) {
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, shareduow.ErrTransactionFailed
	}
	return &AuctionUnitOfWork{
		tx:                tx,
		auctionRepository: repository.NewPostgresAuctionRepository(tx, f.auctionMapper),
		bidRepository:     repository.NewPostgresBidRepository(tx, f.bidMapper),
	}, nil
}
```

`Complete`/`Rollback`、fx 装配、application 层全部零改动。

### 6. 工具链与 CI

- **安装**：`Makefile` 的 `install-libs` 增加 `go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0`（版本钉死，避免生成代码漂移）。
- **生成**：新增 `make sqlc` 目标执行 `sqlc generate`；生成代码提交进 git（消费方无需装 sqlc）。
- **校验**：CI 增加两步——`sqlc vet`（查询静态检查）+ `sqlc generate && git diff --exit-code`（防手改生成代码/忘记重新生成）。
- **lint**：`.golangci.yml` 排除 `**/sqlcgen/**`（生成代码不参与 lint）。

### 7. 迁移的连带修复

改造过程中顺带修复现有实现里的已知问题（均为行为修正，单独列出便于 review）：

1. `bid_repository.go:103` `FindByID` 的 `Scan` 漏了 `max_amount_in_cents` 列（SELECT 了 7 列只 Scan 了 6 列，运行时必然报错）——sqlc 生成代码天然消除此问题。
2. `FindAllPaginated` 两个分支的超长单行 SQL 统一为多行格式。

## 实施顺序

| 阶段 | 内容 | 验证 |
|---|---|---|
| 1 | 引入 sqlc.yaml + 查询文件 + `make sqlc`，生成代码落地 | `sqlc generate` 成功；`go build` 通过 |
| 2 | 改造 users 模块（user / refresh_token 仓储 + mapper，删 entity） | 单测 + `go build` |
| 3 | 改造 auction 模块（auction / bid 仓储 + mapper + UoW factory，删 entity、删 `uow.DBExecutor`） | 单测（含 place_bid 并发路径）+ `go build` |
| 4 | Makefile / CI / golangci 排除规则 / README 技术栈表更新 | CI 全绿 |

先 users 后 auction：users 模块无 UoW、无触发器错误映射，改造面小，先趟通 sqlc 工作流再动核心的竞拍链路。

## 风险与对策

| 风险 | 对策 |
|---|---|
| sqlc 对 goose 迁移里 `-- +goose StatementBegin/End`（PL/pgSQL 触发器）的解析 | sqlc 官方支持 goose 注释；触发器函数体不影响表结构解析。阶段 1 即可验证，若有问题可将 schema 指向独立导出的纯 DDL 文件 |
| `RETURNING *` 使 Create 返回列随迁移自动扩展 | 这是收益而非风险（新列自动进入生成 model）；mapper 是唯一需要跟进的点，编译期即暴露 |
| int64 ↔ uint64 转换散布 | 强制收敛在 mapper 一层；review 时检查仓储代码中不出现裸转换 |
| 生成代码与查询文件不同步 | CI `git diff --exit-code` 兜底 |
| 测试依赖手写 entity | 现有单测 mock 的是 ports 接口（mockery），不触及 entity；仅 mapper 测试需随签名更新 |

## 明确不做

- 不引入 sqlc 管理 schema（迁移仍由 goose 负责，sqlc 只读）。
- 不改 ports 接口、领域模型、application 层。
- 不动 `TxManager` / UoW 的事务语义与隔离级别。
- 不用 sqlc 的 `:copyfrom` / 批量特性（当前无批量场景）。
