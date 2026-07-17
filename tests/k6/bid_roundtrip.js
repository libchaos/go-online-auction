// 主场景:出价受理→WS 通知闭环。
// 每个 bidder VU 自带一条 WS 连接,用自己生成的 Idempotency-Key 关联自己的
// bid.placed 事件,在 WS 回调里算端到端延迟;受理后超时未见事件即计丢失。
// 详见 tasks/prd-nats-auction-core/k6-test-plan.md 第 3、5 节。
import http from 'k6/http';
import { WebSocket } from 'k6/experimental/websockets';
import { setTimeout, clearTimeout, setInterval, clearInterval } from 'k6/experimental/timers';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import {
  bidE2ELatency, bidAcceptLatency, bidEventsLost, bidRoundtripOk, wsSubscribed,
} from './metrics.js';
import {
  BASE, WS, AUCTION_ID, EVENT_TIMEOUT_MS, BID_INTERVAL_MS, SESSION_MS,
  authHeaders, thresholds,
} from './config.js';

export const options = {
  thresholds,
  scenarios: {
    // 稳态:爬升至 N 个并发出价者并持稳。用 --scenario 覆盖选择其一。
    steady_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: Number(__ENV.VUS || 50) },
        { duration: __ENV.HOLD || '5m', target: Number(__ENV.VUS || 50) },
        { duration: '30s', target: 0 },
      ],
      gracefulStop: '10s',
    },
    // 突刺:验证削峰不拒绝(用 arrival-rate 建模每秒出价者到达)。
    spike: {
      executor: 'ramping-arrival-rate',
      startRate: 5,
      timeUnit: '1s',
      preAllocatedVUs: Number(__ENV.MAX_VUS || 300),
      stages: [
        { duration: '10s', target: 20 },
        { duration: '10s', target: Number(__ENV.PEAK || 200) },
        { duration: '30s', target: Number(__ENV.PEAK || 200) },
        { duration: '10s', target: 20 },
      ],
    },
  },
};

export default function () {
  if (!AUCTION_ID) throw new Error('AUCTION_ID is required (see setup.js / config.js)');

  const pending = new Map(); // idempotencyKey -> { sentAt, timer }
  const ws = new WebSocket(`${WS}/ws/auctions/${AUCTION_ID}`);

  ws.onmessage = (e) => {
    let msg;
    try { msg = JSON.parse(e.data); } catch (_) { return; }

    if (msg.type === 'subscription_confirmed') { wsSubscribed.add(true); return; }
    if (msg.event_type !== 'bid.placed') return;

    // 依赖 Task 5.0:bid.placed 载荷需携带 idempotency_key 用于关联。
    const key = msg.data && msg.data.idempotency_key;
    const p = key && pending.get(key);
    if (p) {
      bidE2ELatency.add(Date.now() - p.sentAt);
      bidRoundtripOk.add(true);
      clearTimeout(p.timer);
      pending.delete(key);
    }
  };

  ws.onerror = () => wsSubscribed.add(false);

  ws.onopen = () => {
    const iv = setInterval(() => {
      const key = uuidv4();
      const t0 = Date.now();
      const res = http.post(
        `${BASE}/api/v0/auctions/${AUCTION_ID}/bids`,
        JSON.stringify({ amount_in_cents: 100 + Math.floor(Math.random() * 100000) }),
        { headers: authHeaders({ 'Idempotency-Key': key }) },
      );
      bidAcceptLatency.add(Date.now() - t0);
      check(res, { 'bid accepted 202': (r) => r.status === 202 });
      if (res.status !== 202) return;

      const timer = setTimeout(() => {
        bidEventsLost.add(1);
        bidRoundtripOk.add(false);
        pending.delete(key);
      }, EVENT_TIMEOUT_MS);
      pending.set(key, { sentAt: t0, timer });
    }, BID_INTERVAL_MS);

    // 会话结束:停止出价,给在途事件留出 EVENT_TIMEOUT_MS 收尾,再关闭。
    setTimeout(() => {
      clearInterval(iv);
      setTimeout(() => ws.close(), EVENT_TIMEOUT_MS + 500);
    }, SESSION_MS);
  };
}
