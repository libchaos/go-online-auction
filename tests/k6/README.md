# k6 负载与端到端测试

对照新架构(NATS 竞拍核心 + 通知)的黑盒压测脚本。完整方案见
`tasks/prd-nats-auction-core/k6-test-plan.md`。

## 文件

| 文件 | 作用 |
|---|---|
| `config.js` | 共享 `__ENV` 配置与 thresholds(SLO 闸门) |
| `metrics.js` | 自定义指标(端到端延迟、丢失、闭环率) |
| `setup.js` | 登录 / 创建并 start 竞拍的 helper(端点按实际 API 核对) |
| `bid_roundtrip.js` | 主场景:出价受理→WS 通知闭环(steady / spike) |
| `idempotency.js` | 幂等专项:同一 Idempotency-Key 连发,断言仅一条事件 |
| `watchers.js` | 纯观众扇出:大量只订阅的 WS 客户端 |

## 前置条件

- 安装 k6(需支持 `k6/experimental/websockets`)。
- 起压测环境:`docker-compose up`(NATS JetStream + Postgres),并运行迁移。
- **依赖 Task 5.0**:`bid.placed` 事件载荷需携带 `idempotency_key`,脚本据此关联出价与事件。

## 运行

```bash
# 稳态负载(50 并发出价者,持稳 5 分钟)
BASE_URL=http://localhost:9000 WS_URL=ws://localhost:9000 TOKEN=... AUCTION_ID=42 \
  k6 run --scenario steady_load --env VUS=50 --env HOLD=5m tests/k6/bid_roundtrip.js

# 突刺(削峰验证)
... k6 run --scenario spike --env PEAK=200 tests/k6/bid_roundtrip.js

# 幂等专项
... k6 run --env DUP_COUNT=10 tests/k6/idempotency.js

# 观众扇出
... k6 run --env WATCHERS=500 tests/k6/watchers.js

# 结果留存
... k6 run --out json=results.json tests/k6/bid_roundtrip.js
```

所有目标经 `__ENV` 注入(见 `config.js`),脚本零硬编码。thresholds 失败即 k6 非零退出,可作为 CI 回归闸门(建议非阻断夜间 job)。
