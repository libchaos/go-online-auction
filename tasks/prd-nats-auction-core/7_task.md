# Task 7.0: 移除 Redis

<critical>Read the `prd.md` and `tech-spec.md` files in this folder. If you do not read these files, your task will be considered invalid.</critical>

## Overview

在通知与竞拍核心链路全部切换到 NATS 之后（3.0–6.0），彻底移除 Redis 的代码、依赖、配置、容器编排与文档，达成 PRD 目标“运行不再依赖 Redis”。依赖 3.0、4.0、5.0、6.0。

<requirements>
- 删除 Redis 相关代码：`pkg/redis/*`、`internal/shared/modules/redis/*`、`internal/shared/modules/config/redis.go`。
- 从 `config.Config` 移除 `Redis` 字段（如 1.0 尚保留）。
- `cmd/{auction,websocket,all}.go` 移除 `redis.Module`，改用/确认 `nats.Module`；如新增 `bid-processor` 命令一并接线。
- 移除依赖：`go.mod` 去掉 `github.com/redis/go-redis/v9`；`go mod tidy`。
- 删除 `tests/mocks/mock_UniversalClient.go`。
- 清理 `.env.example` 与 `.env` 中 `REDIS_*`；`docker-compose.yaml` 删除 `redis` 服务与卷。
- 全仓 grep 确认无残留 redis 引用（代码/配置/文档）。
</requirements>

## Subtasks

* [ ] 7.1 删除上述 redis 代码文件与 config 字段
* [ ] 7.2 更新 `cmd/*.go` 模块接线（redis.Module → nats.Module / bid-processor）
* [ ] 7.3 `go.mod` 移除 go-redis；`go mod tidy` 校验无遗留间接依赖
* [ ] 7.4 删除 `tests/mocks/mock_UniversalClient.go`；`make` 重新生成 mock（mockery `all: true`）
* [ ] 7.5 清理 `.env*`、`docker-compose.yaml`
* [ ] 7.6 `grep -ri redis` 全仓核查（含 README，若涉及则在 8.0/文档任务同步）

## Implementation Details

见 `tech-spec.md` 的 “Relevant Files（删除清单）” 与 “Integration Points”。务必在 3.0–6.0 验证通过后执行，避免删除后链路不可用。

## Success Criteria

* 代码库编译、`make lint test` 通过，且不含任何 redis 依赖/引用。
* 应用仅依赖 NATS + Postgres 即可完整运行（出价 + 实时通知）。
* `docker-compose up` 不再启动 redis。

## Relevant Files

* 删除：`pkg/redis/*`、`internal/shared/modules/redis/*`、`internal/shared/modules/config/redis.go`、`tests/mocks/mock_UniversalClient.go`
* 修改：`internal/shared/modules/config/config.go`、`cmd/auction.go`、`cmd/websocket.go`、`cmd/all.go`、`go.mod`、`go.sum`、`docker-compose.yaml`、`.env.example`
