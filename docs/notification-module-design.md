# Notification 模块设计文档

> 状态：设计稿（待评审）。本设计严格对齐本仓库既有约定：六边形 + CQRS + 事件驱动 + DDD + Uber Fx + chi + NATS JetStream + sqlc + goose + PostgreSQL/pgx。
> 所有路径/模式均经代码核查（非凭空假设），可直接作为实施蓝图。

---

## 1. 目标与范围

新建顶层 `notification` 限界上下文，统一承载四类能力：

1. **SSE 实时推送（推送通知）**：`GET /api/v1/notifications/stream` 通过 `text/event-stream` 把新通知实时推送给已登录用户。SSE 即本模块的"推送"通道；浏览器 Web Push（VAPID/FCM）列为可选扩展（见 §14）。
2. **应用内通知中心**：`notifications` 表持久化所有站内通知，支持分页列表、未读过滤、未读数、标记已读/全部已读、删除。SSE 仅做"实时副本"，列表接口才是权威数据源。
3. **通知偏好设置**：`notification_preferences` 表按 `(分类 × 通道)` 控制是否接收；缺省行时返回默认开启策略，跨模块零耦合。
4. **邮件系统**：`EmailPort` 抽象 + `SMTPEmailAdapter` 默认实现，经 NATS 脱机投递，保证至少一次且不重复。

**设计原则**
- 通知模块是**纯消费者**：订阅既有域事件流（`AUCTION_EVENTS`/`DEPOSIT_EVENTS`/`PAYMENT_EVENTS`/`LISTING_EVENTS`），不修改任何生产方模块。
- 偏好、邮件、SSE 三者解耦：偏好决定"发不发"，SSE/邮件是"怎么发"的两个通道。
- 沿用本仓库 Outbox + `Nats-Msg-Id` 幂等模式（参考 `payment` 模块）。

---

## 2. 整体架构

```
                ┌──────────────── 既有域事件流 (NATS JetStream) ────────────────┐
                │ AUCTION_EVENTS  DEPOSIT_EVENTS  PAYMENT_EVENTS  LISTING_EVENTS│
                └───────────────────────────┬──────────────────────────────────┘
                                            │ 订阅 (CreateOrUpdateConsumer, FilterSubject)
                                            ▼
                ┌──────────────── NotificationEventConsumer (application/service) ────────────────┐
                │  解析 source event → 目标 userID + (category, type, title, body, payload)          │
                │  调 NotificationApplicationService.HandleSourceEvent                               │
                └───────────────────────────┬───────────────────────────────────────────────────────┘
                                            ▼
                ┌──────────── NotificationApplicationService (领域编排) ────────────┐
                │  1) 取偏好 PreferenceRepository.GetByUserID (缺省→默认)             │
                │  2) 按偏好决定通道: in_app / email                                   │
                │  3) in_app → NotificationRepository.Insert (幂等键)                 │
                │       └─ 发布 notification.evt.created → NOTIFICATION_EVENTS        │
                │  4) email  → NotificationOutbox.Insert (幂等键)                     │
                │       └─ Relay → notification.evt.email.requested → NOTIFICATION…  │
                └───────────────────────────────────────────────────────────────────┘
                                            │                         │
                        ┌───────────────────┘                         └────────────────────┐
                        ▼                                                     ▼
            ┌── SSE RealtimeHub (infra/http) ──┐              ┌── EmailDispatchConsumer ──┐
            │ NATS consumer: notification.evt. │              │ NATS consumer:            │
            │   .created → 按 userID 路由本地   │              │  notification.evt.email.* │
            │   连接的 SSE 客户端 (registry)   │              │  → EmailPort.Send        │
            └────────────────┬────────────────┘              └───────────────────────────┘
                             ▼                                        ▼
                   浏览器 EventSource                           SMTP / SES / SendGrid …
```

---

## 3. 领域模型 (`domain/`)

### 3.1 枚举 (`domain/enum/`)
遵循命名规范 `Enum{Name}{Value}` + `validate{Name}Enum` + `errors.New`：

- `NotificationCategoryEnum`：`auction` / `bid` / `deposit` / `payment` / `listing` / `system`
- `NotificationChannelEnum`：`in_app` / `email`（/`push` 预留）
- `NotificationTypeEnum`（按分类细分，示例）：
  - bid：`outbid`(被超越) / `bid_won`(中标) / `bid_placed_confirm`(出价确认)
  - auction：`auction_started` / `auction_ending`(即将结束) / `auction_ended`
  - deposit：`deposit_held` / `deposit_released` / `deposit_applied` / `deposit_forfeited`
  - payment：`payment_success` / `withdrawal_requested` / `withdrawal_completed` / `withdrawal_failed`
  - listing：`listing_published`
  - system：`system_security` / `system_account`

### 3.2 聚合 (`domain/model/notification.go`)
不可变值对象按值返回；领域错误 `errs.ErrXxx`。

```go
type Notification struct {
    ID             uint64
    UserID         uint64
    Category       enum.NotificationCategory
    Type           enum.NotificationType
    Title          string
    Body           string
    Payload        map[string]any // 业务上下文（auctionID, amount…），JSONB
    Channels       []enum.NotificationChannel
    IdempotencyKey string
    ReadAt         *time.Time
    CreatedAt      time.Time
}
func (n *Notification) MarkRead(now time.Time)    // 幂等：已读不再改
func (n *Notification) IsRead() bool
```

### 3.3 领域事件 (`domain/event/`)
内部生命周期事件（发布到 `NOTIFICATION_EVENTS`）：`NotificationCreated{UserID, Notification}`、`EmailRequested{UserID, Notification}`。信封复用 `infra/event/envelope` 模式（`ToNotificationCreatedOutboxEvent` 产出 `ports.OutboxEvent{EventID, Subject, Payload, ...}`）。

### 3.4 领域错误 (`domain/errs/errs.go`)
`ErrNotificationNotFound`、`ErrNotificationAccessDenied`（操作他人通知）、`ErrPreferenceInvalid`、`ErrStreamingUnsupported`（非 SSE 客户端）。

---

## 4. 端口 (`ports/`)

```go
type NotificationRepository interface {
    Insert(ctx, Notification) error                 // ON CONFLICT (idempotency_key) DO NOTHING
    ListByUser(ctx, ListNotificationsInput) ([]Notification, int64, error)
    GetUnreadCount(ctx, userID) (int64, error)
    MarkRead(ctx, id, userID) error                 // WHERE id=$1 AND user_id=$2
    MarkAllRead(ctx, userID) error
    Delete(ctx, id, userID) error
    ExistsByID(ctx, id, userID) (bool, error)
}

type PreferenceRepository interface {
    GetByUserID(ctx, userID) (*Preference, error)   // nil → 上层套默认
    Upsert(ctx, Preference) error
}

// 外部动作抽象（参考 payment.AlipayPort）
type EmailPort interface {
    Send(ctx context.Context, msg EmailMessage) error
}
type EmailMessage struct {
    To, Subject, HTMLBody, TextBody string
}

// 可选：Web Push 扩展点（§14），默认不实现
// type PushPort interface { Send(ctx, PushMessage) error }

// 当源事件载荷缺收件人/上下文时的只读解析端口（可选，§9）
type SourceEventResolverPort interface {
    ResolveOutbidUser(ctx, auctionID, winningBidID) (uint64, error)
}
```

---

## 5. 应用层 (`application/`)

### 5.1 Command（CQRS 写）
- `create_notification_command.go` — 内部使用：落库一条通知（含幂等键）。
- `mark_notification_read_command.go` — 标记已读（校验归属）。
- `mark_all_notifications_read_command.go`
- `delete_notification_command.go` — 删除（校验归属）。
- `update_preferences_command.go` — upsert 偏好（校验枚举合法）。

### 5.2 Query（CQRS 读）
- `list_notifications_query.go` — 输入 `UserID, Limit, Offset, Category, UnreadOnly, Type`；输出 `(items, total)`。
- `get_unread_count_query.go`
- `get_preferences_query.go` — 缺省返回 `DefaultPreferences()`。

### 5.3 Service / Consumer（领域编排）
- `notification_event_consumer.go` — 订阅源事件流，反序列化后委托给应用服务（复用 `payment` 的 `Start/Stop` + `fx.Hook` 生命周期模式）。
- `notification_application_service.go` — `HandleSourceEvent`：映射 → 取偏好 → 落库 in_app + 发布 `notification.evt.created`；email 通道 → 写 outbox。
- `realtime_hub.go` — **SSE 版 Hub**，镜像 `deposit/infra/websocket/deposit_hub.go` 的 `UserSubscriberRegistry` + `EventConsumer` 扇出模式（见 §6）。
- `email_dispatch_consumer.go` — 消费 `notification.evt.email.*`，调 `EmailPort.Send`（幂等键去重）。

---

## 6. SSE 实时推送（核心）

**复用既有 websocket Hub 范式**（`deposit/infra/websocket/deposit_hub.go`）：内存 `UserSubscriberRegistry` + NATS `EventConsumer` 按 `userID` 路由。SSE 与 WS 的区别仅在传输层（HTTP 分块流 vs WebSocket 帧）。

`infra/http/chi/handler` 的 SSE 端点：

```go
func (h *NotificationHandler) StreamNotifications(w http.ResponseWriter, r *http.Request) {
    // SSE 鉴权：?token= 优先，回退 Authorization 头（EventSource 不能带请求头）
    token := r.URL.Query().Get("token")
    if token == "" {
        if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
            token = strings.TrimPrefix(h, "Bearer ")
        }
    }
    claims, err := h.tokenVerifier.Verify(token)
    if err != nil {
        response.Error(w, authn.ErrUnauthorized)
        return
    }
    ctx := authn.WithClaims(r.Context(), claims)

    flusher, ok := w.(http.Flusher)
    if !ok { response.Error(w, errs.ErrStreamingUnsupported); return }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("X-Accel-Buffering", "no") // 关闭反向代理缓冲
    w.WriteHeader(http.StatusOK)

    hub := h.realtimeHub
    done := make(chan struct{})
    send := make(chan []byte, 16)
    hub.Register(claims.UserID, send)
    defer hub.Unregister(claims.UserID, send)

    for {
        select {
        case <-r.Context().Done():
            return
        case data := <-send:
            _, _ = fmt.Fprintf(w, "id: %s\n", extractID(data))
            _, _ = fmt.Fprintf(w, "event: notification\n")
            _, _ = fmt.Fprintf(w, "data: %s\n\n", data)
            flusher.Flush()
        }
    }
}
```

**RealtimeHub（扇出）** 与 `deposit_hub.go` 同构：
- `registry *UserSubscriberRegistry`（`map[uint64]map[chan []byte]struct{}`，`Add/Remove/PublishToUser`）。
- NATS 消费者订阅 `notification.evt.created`（或 `notification.evt.>`），按 payload 中的 `user_id` 调 `registry.PublishToUser`。
- 采用 **queue group**：多实例各订阅，NATS 保证一条消息只投递到一个实例，该实例只推送给本机连接的该用户 SSE 客户端 → 天然支持水平扩容。

**客户端行为**：连上 SSE 后先 `GET /notifications` 拉历史（初始状态），之后仅收增量。

**SSE 鉴权（已确认）**：经核查 `authn.RequireAuth` 仅解析 `Authorization` 头（Bearer，见 `internal/shared/modules/authn/middleware.go`），浏览器 `EventSource` 无法设置请求头。故 SSE 端点**不挂 `middleware.RequireAuth`**，改为在 handler 内注入 `authn.TokenVerifier`（与 `authn.NewMiddleware` 同一实例），按序取 token：① `r.URL.Query().Get("token")`；② 回退 `Authorization` 头；校验通过后 `authn.WithClaims(r.Context(), claims)` 注入上下文。前端用 `new EventSource('/api/v1/notifications/stream?token='+jwt)` 连接。`authn.TokenVerifier` 已作为接口在 fx 图提供（`users/module.go` 的 `fx.As(new(authn.TokenVerifier))`），SSE handler 直接注入即可，无需改动 authn 模块。

---

## 7. 邮件系统 (`infra/email/`)

```go
// SMTPEmailAdapter 默认实现，配置来自 config.Config.Email
type SMTPEmailAdapter struct { cfg config.Email }
func (a *SMTPEmailAdapter) Send(ctx context.Context, msg ports.EmailMessage) error {
    // net/smtp 或轻量库；TLS；From= cfg.From；失败返回 error 触发 NATS 重投
}
```

- **投递保证**：邮件请求先写 `notification_outbox`（与通知落库解耦），由 `notification` 模块的 Relay 发布 `notification.evt.email.requested`；`EmailDispatchConsumer` 用 `AckExplicitPolicy` 消费，`Send` 成功才 Ack。幂等键 `source_event_id:user_id:email` 去重，避免重复发信。
- **可插拔**：`fx.As(new(ports.EmailPort))` 注入；将来可换 SES/SendGrid/阿里云 DirectMail 适配器，不动领域层。

---

## 8. 通知偏好设置

```sql
CREATE TABLE notification_preferences (
    user_id    BIGINT PRIMARY KEY,
    preferences JSONB NOT NULL,   -- {"auction":{"in_app":true,"email":false}, "payment":{"in_app":true,"email":true}, ...}
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- `PreferenceRepository.GetByUserID` 返回 `nil` 时，应用层套 `DefaultPreferences()`：
  - `in_app`：全分类默认 `true`（站内中心总是可用）。
  - `email`：默认 `true` 仅 `payment`/`deposit`/`system`（重要），`auction`/`bid`/`listing` 默认 `false`（避免邮件骚扰）。可在 `domain/model/preference.go` 集中定义并单测。
- 不监听 users 模块创建事件（避免跨模块写），缺省即默认，用户首次 GET 即拿到合理默认值。

---

## 9. 源事件 → 通知映射表

`NotificationApplicationService` 维护映射。收件人解析：
- 经核查 `auction/domain/event/bid_placed_event.go` 的 `BidPlacedEvent` 仅含 `BidID`/`AuctionID`/`UserID`(新出价者)/`Amount`，**不含上一位出价者**。因此"被超越"通知必须走只读 `SourceEventResolverPort`（即 `AuctionReadPort`，查询该拍品上一手最高出价用户）。该端口为只读依赖，不反向修改 auction 模块。
- 其余事件（deposit/payment/listing/system）收件人即事件主体用户，直接取自 payload。

| 源事件 (subject) | 分类 | 类型 | 收件人 | 标题模板 | 通道默认 |
|---|---|---|---|---|---|
| `auction.evt.bid_placed` | bid | `outbid` | 经 `AuctionReadPort` 解析的上一位最高出价者 | "您在拍品 #{auctionID} 的出价已被超越" | in_app |
| `auction.evt.auction_ended` | auction | `auction_ended` | 参与过该拍品者 / 关注者 | "拍品 #{auctionID} 已结束" | in_app |
| `auction.evt.auction_started` | auction | `auction_started` | 关注者 | "您关注的拍品已开始" | in_app |
| `deposit.evt.deposit_held` | deposit | `deposit_held` | 该用户 | "保证金已冻结 ¥{amount}" | in_app+email |
| `deposit.evt.deposit_released` / `applied` / `forfeited` | deposit | 对应 | 该用户 | "保证金已释放/抵扣/罚没" | in_app+email |
| `payment.evt.payment_success` | payment | `payment_success` | 该用户 | "充值成功 ¥{amount}" | in_app+email |
| `payment.>.withdrawal_*` | payment | `withdrawal_*` | 该用户 | "提现已受理/完成/失败" | in_app+email |
| `listing.evt.sku.published` | listing | `listing_published` | 关注卖家者 | "您关注的卖家上新" | in_app |
| 系统事件（预留） | system | `system_security`/`system_account` | 该用户 | 安全/账户提醒 | in_app+email |

---

## 10. HTTP API（`infra/http/chi/`）

路由（`notification_router.go`，全部 `middleware.RequireAuth`）：

| Method | Path | 说明 |
|---|---|---|
| GET | `/api/v1/notifications/stream` | **SSE** 实时推送 |
| GET | `/api/v1/notifications` | 列表：`limit`/`offset`/`category`/`type`/`unread`(bool) |
| GET | `/api/v1/notifications/unread-count` | 未读数 |
| POST | `/api/v1/notifications/{id}/read` | 标记已读（校验归属） |
| POST | `/api/v1/notifications/read-all` | 全部已读 |
| DELETE | `/api/v1/notifications/{id}` | 删除（校验归属） |
| GET | `/api/v1/notifications/preferences` | 取偏好 |
| PUT | `/api/v1/notifications/preferences` | 更新偏好 |

- 响应经 `response.JSON` 包 `{data:...}`；分页响应 `dto.NotificationListResponse{TotalCount,Limit,Offset,Items}`，错误经 `httperrs.MapDomainError`。
- DTO/mapper 在 `infra/http/dto` + `infra/mapper`，与既有 `payment_handler` 同构。

---

## 11. 持久化（迁移 + sqlc）

新增迁移（当前最大 `000021`，接续编号）：

- `000022_create_notifications_table.sql`
- `000023_create_notification_preferences_table.sql`
- `000024_create_notification_outbox_table.sql`

`000022` 示例（对齐 `payments_outbox` 风格）：

```sql
-- +goose Up
CREATE TABLE notifications (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT       NOT NULL,
    category        VARCHAR      NOT NULL,
    type            VARCHAR      NOT NULL,
    title           VARCHAR      NOT NULL,
    body            TEXT         NOT NULL,
    payload         JSONB,
    channels        VARCHAR[]    NOT NULL,
    idempotency_key VARCHAR      NOT NULL,
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_notifications_idempotency UNIQUE (idempotency_key)
);
CREATE INDEX idx_notifications_user_unread ON notifications (user_id, created_at DESC)
    WHERE read_at IS NULL;
CREATE INDEX idx_notifications_user_created ON notifications (user_id, created_at DESC);
-- +goose Down
DROP TABLE IF EXISTS notifications;
```

`sqlc.yaml` 追加 notification 段：

```yaml
  - engine: "postgresql"
    schema: "migrations"
    queries: "internal/modules/notification/infra/query"
    gen:
      go:
        package: "sqlcgen"
        out: "internal/modules/notification/infra/sqlcgen"
        sql_package: "pgx/v5"
        emit_pointers_for_null_types: true
        omit_unused_structs: true
        overrides:
          - db_type: "timestamptz"
            go_type: "time.Time"
          - db_type: "timestamptz"
            nullable: true
            go_type: { type: "time.Time", pointer: true }
```

查询（`infra/query/notifications.sql` / `notification_preferences.sql` / `notification_outbox.sql`）沿用命名约定 `InsertNotification`/`:many ListNotificationsByUser`/`:one GetUnreadNotificationCount`/`:exec MarkNotificationRead`/`:execrows …`。

---

## 12. NATS 流与主题

`internal/shared/modules/nats/streams.go` 追加：

```go
StreamNotificationEvents  = "NOTIFICATION_EVENTS"
SubjectNotificationEvents = "notification.evt.>"
```

并在 `CreateOrUpdateStreams` 的 `configs` 中加入 `NOTIFICATION_EVENTS`（LimitsPolicy + FileStorage + Duplicates 窗口，与 `AUCTION_EVENTS` 一致）。

- 源事件订阅：在 `NotificationEventConsumer.Start` 中对每个源流 `CreateOrUpdateConsumer`（durable name 如 `notification-auction-events`，`FilterSubject` 对应 `auction.evt.*`/`deposit.evt.*`/`payment.>`/`listing.evt.>`）。**只读消费，不改动既有流。**
- 内部事件：`notification.evt.created`（SSE Hub 消费）、`notification.evt.email.requested`（邮件消费）。

---

## 13. Fx 装配（`module.go`）

对齐 `payment/module.go`：

```go
var Module = fx.Module("notification",
    fx.Provide(mapper.NewNotificationMapper),
    fx.Provide(fx.Annotate(repository.NewPostgresNotificationRepository, fx.As(new(ports.NotificationRepository)))),
    fx.Provide(fx.Annotate(repository.NewPostgresPreferenceRepository, fx.As(new(ports.PreferenceRepository)))),
    fx.Provide(fx.Annotate(email.NewSMTPEmailAdapter, fx.As(new(ports.EmailPort)))),
    fx.Provide(command.NewCreateNotificationCommand),
    fx.Provide(command.NewMarkNotificationReadCommand),
    fx.Provide(command.NewMarkAllNotificationsReadCommand),
    fx.Provide(command.NewDeleteNotificationCommand),
    fx.Provide(command.NewUpdatePreferencesCommand),
    fx.Provide(query.NewListNotificationsQuery),
    fx.Provide(query.NewGetUnreadCountQuery),
    fx.Provide(query.NewGetPreferencesQuery),
    fx.Provide(service.NewNotificationApplicationService),
    fx.Provide(service.NewNotificationEventConsumer),
    fx.Provide(service.NewRealtimeHub),
    fx.Provide(service.NewEmailDispatchConsumer),
    fx.Provide(handler.NewNotificationHandler),
    fx.Provide(notificationoutbox.NewRelay),
)

func RegisterNotificationRoutes(server *httpserver.Server, h *handler.NotificationHandler, m *authn.Middleware) {
    router.RegisterNotificationRoutes(server, h, m)
}
func RegisterNotificationConsumers(lc fx.Lifecycle, src *service.NotificationEventConsumer,
    hub *service.RealtimeHub, email *service.EmailDispatchConsumer, logger logger.Logger) { /* OnStart/OnStop */ }
func RegisterNotificationOutboxRelay(lc fx.Lifecycle, relay *notificationoutbox.Relay, logger logger.Logger) { /* OnStart/OnStop */ }
```

`cmd/all.go` 的 `fx.New(... fx.Invoke(..., notification.Module, notification.RegisterNotificationRoutes, notification.RegisterNotificationConsumers, notification.RegisterNotificationOutboxRelay, ...))`。（`cmd/auction.go` 独立子命令同步。）

---

## 14. 非目标 / 可选扩展

- **Web Push（VAPID/FCM）**：新增 `PushPort` + 订阅管理（`push_subscriptions` 表 + `POST /notifications/push-subscribe`），`NotificationApplicationService` 在 `push` 通道开启时调用。SSE 与 Web Push 互补：SSE 覆盖在线同标签页，Web Push 覆盖离线/跨标签页。
- **通知聚合/摘要**：邮件每日摘要（timer 消费者）未来可加。
- **多端已读同步**：以 `notifications.read_at` 为权威，SSE 推送 `read` 事件即可让其他标签页同步。

---

## 15. 待确认决策点

1. **SSE 鉴权方式** ✅ 已确认：因 `authn.RequireAuth` 仅解析 `Authorization` 头，`EventSource` 无法带请求头，SSE 端点改用 `?token=`（回退 `Authorization` 头），复用注入的 `authn.TokenVerifier` + `authn.WithClaims` 注入上下文。
2. **`bid_placed` 是否含 `previous_bidder_id`** ✅ 已确认：不含。经核查 `BidPlacedEvent` 仅含新出价者 `UserID`。"被超越"通知须经只读 `AuctionReadPort`（`SourceEventResolverPort`）解析上一手最高出价用户，不修改 auction 模块。
3. **邮件默认开关**：默认 email 仅 `payment/deposit/system` 开启是否合理？还是全开？（建议维持现状：重要类开、噪音类关）
4. **email 投递实现**：Outbox + Relay（与 `payment` 一致，推荐，保证不丢）vs 直接 NATS 发布（简化）。本设计采用前者。

---

## 16. 实施步骤（建议分期）

1. **脚手架 + 持久化**：迁移 `000022/23/24` + sqlc 段 + repository + mapper + ports + `sharednats` 流；`make sqlc`。
2. **偏好 + 中心读写**：command/query + handler/router（除 SSE）+ `module.go` 接线 + `cmd` 注册。
3. **SSE 实时**：`RealtimeHub` + `notification.evt.created` 发布 + SSE handler/路由；单元测扇出。
4. **事件订阅 + 编排**：`NotificationEventConsumer` + `NotificationApplicationService` + 映射表；先接 `payment.>`/`deposit.evt.*`，再接 `auction.evt.*`/`listing.evt.>`。
5. **邮件**：`EmailPort` + `SMTPEmailAdapter` + outbox + `EmailDispatchConsumer` + 配置项。
6. **验证门禁**：`go build/vet/test/lint/nilaway` 全绿；`make update-mocks`；testcontainers 集成测试覆盖"源事件 → 通知落库 + SSE 推送 + 邮件请求"。

---

## 17. 测试策略

- **Unit**：command/query 用 `tests/mocks` 的 repository mock（参考 payment 测试，外部 `package command_test`/`query_test`）；consumer/服务用白盒 `package service`；偏好默认策略单测。
- **SSE Hub**：内存 `UserSubscriberRegistry` 扇出单测（模拟 NATS 推送 → 校验目标用户客户端收到、非目标不收）。
- **Email**：mock `EmailPort` 验证幂等键去重、失败重投。
- **Integration**（testcontainers，沿用既有模式）：发布源域事件 → 断言 `notifications` 落库正确、未读计数、SSE 收到、邮件请求进入 outbox。

---

## 18. 幂等 / 顺序 / 错误处理要点

- **幂等**：`notifications.idempotency_key = source_event_id:user_id:channel` UNIQUE；`Insert … ON CONFLICT DO NOTHING`；邮件同理。NATS `Duplicates` 窗口 + `Nats-Msg-Id` 双重保险（沿用全局约定）。
- **顺序**：列表按 `created_at DESC`；同一用户多连接顺序由 NATS 投递保证。
- **归属校验**：所有写操作（`MarkRead`/`Delete`）`WHERE user_id=$2`，否则 `ErrNotificationAccessDenied`。
- **SSE 断线**：客户端重连后 `GET /notifications` 拉全量，SSE 仅补增量；Hub `defer Unregister` 防泄漏。
- **nilaway**：生成 mock 包已在 `Makefile` 排除（`auction/tests/mocks`），新增 UoW mock 不受影响。
