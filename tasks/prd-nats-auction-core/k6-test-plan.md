# k6 负载与端到端测试方案 — NATS 竞拍核心与通知

> 参考本目录 `prd.md` / `tech-spec.md`。本方案通过系统**公开接口**（HTTP + WebSocket）验证新架构在负载下的行为，与 Task 8.0 的 Go 内嵌 NATS 集成测试**互补**（后者验证 JetStream 内部：消费顺序、幂等、DLQ、Stream 堆积）。

## 1. 目标与测试边界

k6 从**外部黑盒**视角回答四个问题，直接对应 PRD 的 Objectives：

| PRD 目标 | k6 验证方式 |
|---|---|
| 可靠性/不丢消息（Obj.1） | 统计“出价被 202 受理”但其 `bid.placed` 事件在超时内**未经 WS 送达**的比例 = 事件丢失率 |
| 可扩展/削峰（Obj.2） | 高并发出价下，HTTP 接受层 P95 延迟与错误率是否**与落库解耦**保持平稳 |
| 端到端体验 | 测量**受理→WS 通知**的端到端延迟（新架构最关键的新指标） |
| 正确性/幂等（Obj.5, FR 1.2） | 同一 `Idempotency-Key` 重复提交，最终 WS 仅出现一条对应 `bid.placed` |

**边界（k6 不做的）**：NATS 消费顺序/公平、DLQ 路径、Stream 积压告警——这些无外部可观测面，由 Task 8.0 Go 集成测试 + 服务端 Prometheus 指标覆盖。k6 不直接连 NATS。

## 2. 自定义指标（Custom Metrics）

```javascript
import { Trend, Counter, Rate } from 'k6/metrics';

// 受理(HTTP 202)→ 收到对应 bid.placed(WS) 的端到端延迟(ms)
export const bidE2ELatency = new Trend('bid_e2e_latency_ms', true);
// HTTP 出价受理延迟(仅接受层，衡量与 DB 解耦效果)
export const bidAcceptLatency = new Trend('bid_accept_latency_ms', true);
// 已受理但在超时窗口内未收到 WS 事件 → 计为丢失
export const bidEventsLost = new Counter('bid_events_lost');
// 成功闭环(受理且收到事件)的比率
export const bidRoundtripOk = new Rate('bid_roundtrip_ok');
// WS 订阅确认成功率
export const wsSubscribed = new Rate('ws_subscribed');
```

## 3. 场景矩阵（Scenarios / Executors）

用 k6 `scenarios` 表达不同压测形态，均针对**已开始（started）**的竞拍：

| 场景 | executor | 目的 | 关键参数 |
|---|---|---|---|
| `smoke` | `shared-iterations` | 冒烟：链路连通、脚本正确 | 1 VU，几次迭代 |
| `steady_load` | `ramping-vus` | 稳态负载：验证 SLO | 爬升至 N VU 持稳 5–10m |
| `stress` | `ramping-arrival-rate` | 逼近拐点：找吞吐上限与退化点 | 出价速率阶梯上升 |
| `spike` | `ramping-arrival-rate` | 突刺：验证削峰不拒绝 | 瞬间拉高再回落 |
| `soak` | `constant-vus` | 长稳：内存/连接/积压泄漏 | 中等负载 30–60m |

> 出价负载优先用 **arrival-rate**（按“每秒出价数”建模），比固定 VU 更贴近真实竞拍洪峰；WS 订阅者用 VU 建模“在线观众数”。

## 4. Thresholds（SLO 闸门）

阈值中的**具体数值**依赖 PRD Open Questions（P95 延迟目标、峰值吞吐）——先参数化占位，量化目标确定后收紧：

```javascript
export const options = {
  thresholds: {
    // 接受层与 DB 解耦：受理应始终很快
    'bid_accept_latency_ms': ['p(95)<200'],
    // 端到端(含异步落库+通知)——占位，待 PRD 量化
    'bid_e2e_latency_ms': ['p(95)<1500', 'p(99)<3000'],
    // 可靠性：事件丢失必须极低(理想为 0)
    'bid_events_lost': ['count<1'],
    'bid_roundtrip_ok': ['rate>0.999'],
    'ws_subscribed': ['rate>0.99'],
    // HTTP 层错误率
    'http_req_failed': ['rate<0.01'],
  },
};
```

任一 threshold 失败 → k6 退出码非 0 → CI 判定回归失败（作为质量闸门）。

## 5. 关键脚本模式：把 202 与 WS 事件关联起来

核心难点是**跨协议关联**：出价通过 HTTP 发出、结果从 WS 回来。方案是每个 bidder VU 自带一条 WS 连接，用自己生成的 `Idempotency-Key` 匹配自己的 `bid.placed`，在 WS 消息回调里算端到端延迟；受理后超时未见事件即计丢失。

```javascript
import http from 'k6/http';
import { WebSocket } from 'k6/experimental/websockets';
import { setTimeout, clearTimeout } from 'k6/experimental/timers';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import {
  bidE2ELatency, bidAcceptLatency, bidEventsLost,
  bidRoundtripOk, wsSubscribed,
} from './metrics.js';

const BASE = __ENV.BASE_URL || 'http://localhost:9000';
const WS   = __ENV.WS_URL   || 'ws://localhost:9000';
const AUCTION_ID = __ENV.AUCTION_ID; // 由准备脚本创建并 start 的竞拍
const EVENT_TIMEOUT_MS = Number(__ENV.EVENT_TIMEOUT_MS || 5000);

export default function () {
  const pending = new Map(); // idempotencyKey -> { sentAt, timer }
  const ws = new WebSocket(`${WS}/ws/auctions/${AUCTION_ID}`);

  ws.onmessage = (e) => {
    const msg = JSON.parse(e.data);
    if (msg.type === 'subscription_confirmed') { wsSubscribed.add(true); return; }
    if (msg.event_type !== 'bid.placed') return;

    const key = msg.data?.idempotency_key;
    const p = key && pending.get(key);
    if (p) {
      bidE2ELatency.add(Date.now() - p.sentAt);
      bidRoundtripOk.add(true);
      clearTimeout(p.timer);
      pending.delete(key);
    }
  };

  ws.onopen = () => {
    // 在本 VU 的生命周期内按节奏出价
    const iv = setInterval(() => {
      const key = uuidv4();
      const t0 = Date.now();
      const res = http.post(
        `${BASE}/api/v0/auctions/${AUCTION_ID}/bids`,
        JSON.stringify({ amount_in_cents: 100 + Math.floor(Math.random() * 100000) }),
        { headers: { 'Content-Type': 'application/json',
                     'Idempotency-Key': key,
                     'Authorization': `Bearer ${__ENV.TOKEN}` } },
      );
      bidAcceptLatency.add(Date.now() - t0);
      check(res, { 'bid accepted 202': (r) => r.status === 202 });
      if (res.status !== 202) return;

      // 受理后启动丢失计时器：超时未见事件即计丢失
      const timer = setTimeout(() => {
        bidEventsLost.add(1);
        bidRoundtripOk.add(false);
        pending.delete(key);
      }, EVENT_TIMEOUT_MS);
      pending.set(key, { sentAt: t0, timer });
    }, 500); // 每 VU 每 500ms 一次出价

    // VU 会话时长后收尾
    setTimeout(() => { clearInterval(iv); ws.close(); }, Number(__ENV.SESSION_MS || 30000));
  };

  ws.onerror = () => wsSubscribed.add(false);
}
```

> 说明：
> - 采用 `k6/experimental/websockets`（异步、单 VU 可同时持 WS 并发 HTTP）；模块 API 需按所用 k6 版本核对。
> - **前提**：Task 5.0 需让 `bid.placed` 事件载荷携带 `idempotency_key`（用于关联），若未带则改用 `bid_id` + 服务端回执做二级关联。
> - 纯观众负载可用只订阅不出价的精简版脚本（衡量扇出广度与掉线率）。

## 6. 测试数据与环境准备（setup）

- `setup()` 或独立准备脚本：注册/登录测试用户获取 `TOKEN` → 创建 auction → `start` → 导出 `AUCTION_ID`。
- 环境隔离：指向专用压测环境（独立 Postgres + NATS），避免污染开发数据；用 `docker-compose` 起 NATS(JetStream)+Postgres。
- 参数化：所有目标(`BASE_URL`/`WS_URL`/`TOKEN`/`AUCTION_ID`/速率/时长)经 `__ENV` 注入，脚本零硬编码。
- 幂等专项：单独一个小脚本用**同一** `Idempotency-Key` 连发 N 次，断言 WS 仅收到 1 条对应事件、DB(或列表接口)仅 1 条 bid。

## 7. 运行方式与 CI

```bash
# 本地稳态负载
BASE_URL=http://localhost:9000 WS_URL=ws://localhost:9000 \
TOKEN=... AUCTION_ID=42 \
k6 run --env SESSION_MS=60000 tests/k6/bid_roundtrip.js

# 指定场景与输出
k6 run --scenario steady_load --out json=results.json tests/k6/bid_roundtrip.js
```

- **CI**：作为非阻断的夜间/按需 job（负载测试耗时且需环境），thresholds 失败即标记回归。可用 official `grafana/k6` GitHub Action 或容器执行。
- **结果留存**：`--out json` 或 k6 → Prometheus remote-write，进 Grafana 看趋势。

## 8. 与服务端指标关联

把 k6 的黑盒结论与 Task 8.0 的服务端 Prometheus 指标对照，定位瓶颈层：

- `bid_events_lost > 0` → 对照消费者 `redelivered` / `pending` 与 `auction_bid_commands_processed_total{result=dlq}`，判断是丢失还是滞留/进 DLQ。
- `bid_e2e_latency` 高但 `bid_accept_latency` 低 → 瓶颈在 Processor/DB，而非接受层（证明削峰解耦生效）。
- spike 场景下 `http_req_failed` 保持低 + 命令流 `pending` 短时上升后回落 → 削峰达标。

## 9. 相关文件

* 新增：`tests/k6/metrics.js`（自定义指标）、`tests/k6/bid_roundtrip.js`（主场景）、`tests/k6/idempotency.js`（幂等专项）、`tests/k6/watchers.js`（纯观众扇出）
* 新增：`tests/k6/options/*.js` 或 `--scenario` 配置；准备脚本 `tests/k6/setup.js`
* 参考：本目录 `tech-spec.md`（Monitoring and Observability 指标命名）、`prd.md`（Objectives 与 Open Questions 的量化目标）
* 依赖：k6 ≥ 支持 `k6/experimental/websockets` 的版本；可选 xk6 扩展用于 Prometheus 输出
