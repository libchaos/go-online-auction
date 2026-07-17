# Task 1.0: NATS 基础封装与 fx 接入

<critical>Read the `prd.md` and `tech-spec.md` files in this folder. If you do not read these files, your task will be considered invalid.</critical>

## Overview

搭建 NATS.io（JetStream）的连接与接入基础，作为后续所有任务的地基：薄封装 `pkg/nats`、fx 共享模块、配置项，以及在启动时声明式创建 JetStream Stream。本任务不改动任何出价/通知业务逻辑。

<requirements>
- 新增 `pkg/nats`：提供连接与 JetStream 上下文，支持 URL、连接名、可选 creds、TLS、重连、去重窗口配置。
- 新增 `internal/shared/modules/nats`：fx Provider 暴露 `*nats.Conn` 与 `jetstream.JetStream`，生命周期钩子内幂等创建/更新 Stream，并优雅关闭。
- 新增 `internal/shared/modules/config/nats.go` 并在 `config.Config` 中以 `NATS` 字段替换 `Redis`（本任务先新增，Redis 移除在 7.0）。
- 声明式创建 3 个 Stream：`AUCTION_COMMANDS`(WorkQueue,file,subj `auction.cmd.bid.*`)、`AUCTION_EVENTS`(Limits,file,subj `auction.evt.*`)、`AUCTION_DLQ`(Limits,file,subj `auction.dlq.*`)。
- `docker-compose.yaml` 新增 `nats:2.10-alpine`（`-js -sd /data` + 数据卷）；`.env.example` 增加 `NATS_*`。
- 引入依赖 `github.com/nats-io/nats.go`（含 `jetstream` 包）。
</requirements>

## Subtasks

* [x] 1.1 新增 `pkg/nats/config.go`（`Config`）与 `pkg/nats/nats.go`（`New`：连接 + JetStream + 健康检查）
* [x] 1.2 新增 `internal/shared/modules/config/nats.go` 并在 `config.go` 增加 `NATS` 字段（mapstructure squash）
* [x] 1.3 新增 `internal/shared/modules/nats/{nats.go,module.go}`：fx Provider + OnStop 关闭
* [x] 1.4 新增 `streams.go`：`CreateOrUpdateStream` 幂等声明 3 个 Stream，并在 OnStart 调用
* [x] 1.5 更新 `docker-compose.yaml`（nats 服务 + 卷）与 `.env.example`（`NATS_URL` 等）
* [x] 1.6 `go get github.com/nats-io/nats.go`；`go mod tidy`
* [x] 1.7 单测：`pkg/nats` 配置校验；Stream 声明幂等性（内嵌 NATS，见 8.0 可先起草）

## Implementation Details

见 `tech-spec.md` 的 “System Architecture / Component Overview”“Core Interfaces（`pkg/nats`）”“Subject 与 Stream 设计” 与 “Integration Points”。Subject 采用点式令牌（`auction.cmd.bid.{id}` / `auction.evt.{id}` / `auction.dlq.{id}`）。

## Success Criteria

* 应用以 `nats.Module` 启动后成功连接 NATS 并幂等创建 3 个 Stream（重复启动不报错）。
* 停止时干净关闭连接，无 goroutine/连接泄漏。
* `docker-compose up` 提供可用的启用 JetStream 的 NATS。
* `make lint` 通过。

## Relevant Files

* 新增：`pkg/nats/nats.go`、`pkg/nats/config.go`
* 新增：`internal/shared/modules/nats/{nats.go,module.go}`、`internal/shared/modules/nats/streams.go`
* 新增：`internal/shared/modules/config/nats.go`；修改：`internal/shared/modules/config/config.go`
* 修改：`docker-compose.yaml`、`.env.example`、`go.mod`、`go.sum`
