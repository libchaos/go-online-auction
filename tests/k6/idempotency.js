// 幂等专项:同一 Idempotency-Key 连发 N 次,验证最终只产生一条出价、
// WS 只回推一条对应 bid.placed。对应 PRD FR 1.2 与 tech-spec 三层幂等。
// 详见 tasks/prd-nats-auction-core/k6-test-plan.md 第 6 节。
import http from 'k6/http';
import { WebSocket } from 'k6/experimental/websockets';
import { setTimeout } from 'k6/experimental/timers';
import { check } from 'k6';
import { Counter } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { BASE, WS, AUCTION_ID, authHeaders } from './config.js';

const dupEventsReceived = new Counter('idempotency_dup_events'); // 同一 key 收到 >1 条即异常

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    'idempotency_dup_events': ['count<1'], // 必须为 0
    'checks': ['rate>0.99'],
  },
};

const N = Number(__ENV.DUP_COUNT || 10);
const WAIT_MS = Number(__ENV.WAIT_MS || 6000);

export default function () {
  if (!AUCTION_ID) throw new Error('AUCTION_ID is required');

  const key = uuidv4();
  let eventsForKey = 0;
  const ws = new WebSocket(`${WS}/ws/auctions/${AUCTION_ID}`);

  ws.onmessage = (e) => {
    let msg;
    try { msg = JSON.parse(e.data); } catch (_) { return; }
    if (msg.event_type === 'bid.placed' && msg.data && msg.data.idempotency_key === key) {
      eventsForKey += 1;
      if (eventsForKey > 1) dupEventsReceived.add(1);
    }
  };

  ws.onopen = () => {
    // 用同一 key + 同一金额连发 N 次(去重窗口内)。
    let accepted = 0;
    for (let i = 0; i < N; i++) {
      const res = http.post(
        `${BASE}/api/v0/auctions/${AUCTION_ID}/bids`,
        JSON.stringify({ amount_in_cents: 12345 }),
        { headers: authHeaders({ 'Idempotency-Key': key }) },
      );
      check(res, { 'each accepted 202': (r) => r.status === 202 });
      if (res.status === 202) accepted += 1;
    }
    check(null, { 'all duplicates accepted (idempotent, no error)': () => accepted === N });

    // 等待事件稳定后收尾并断言恰好 1 条。
    setTimeout(() => {
      check(null, { 'exactly one bid.placed for key': () => eventsForKey === 1 });
      ws.close();
    }, WAIT_MS);
  };
}

// 收尾:也应通过列表接口断言该 auction 下该 key 仅一条 bid(白盒交叉验证)。
// 可在 teardown 或独立步骤调用 GET /api/v0/auctions/{id} 校验最高价/出价数。
