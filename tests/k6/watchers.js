// 纯观众扇出场景:大量只订阅、不出价的 WS 客户端,衡量通知扇出广度、
// 订阅成功率与掉线率(模拟热门竞拍的围观流量)。可与 bid_roundtrip 同时跑,
// 或先起 watchers 再用另一进程注入出价负载。
// 详见 tasks/prd-nats-auction-core/k6-test-plan.md 第 3 节。
import { WebSocket } from 'k6/experimental/websockets';
import { setTimeout } from 'k6/experimental/timers';
import { Counter, Rate } from 'k6/metrics';
import { WS, AUCTION_ID, SESSION_MS } from './config.js';

const eventsReceived = new Counter('watcher_events_received');
const subscribed = new Rate('ws_subscribed');
const disconnectErrors = new Counter('watcher_disconnect_errors');

export const options = {
  scenarios: {
    watchers: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: Number(__ENV.WATCHERS || 500) },
        { duration: __ENV.HOLD || '5m', target: Number(__ENV.WATCHERS || 500) },
        { duration: '20s', target: 0 },
      ],
    },
  },
  thresholds: {
    'ws_subscribed': ['rate>0.99'],
    'watcher_disconnect_errors': ['count<1'],
  },
};

export default function () {
  if (!AUCTION_ID) throw new Error('AUCTION_ID is required');

  const ws = new WebSocket(`${WS}/ws/auctions/${AUCTION_ID}`);

  ws.onmessage = (e) => {
    let msg;
    try { msg = JSON.parse(e.data); } catch (_) { return; }
    if (msg.type === 'subscription_confirmed') { subscribed.add(true); return; }
    if (msg.event_type) eventsReceived.add(1);
  };

  ws.onerror = () => { subscribed.add(false); disconnectErrors.add(1); };

  ws.onopen = () => {
    setTimeout(() => ws.close(), SESSION_MS);
  };
}
