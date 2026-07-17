// 共享配置与 thresholds — 全部经 __ENV 注入,脚本零硬编码。
// 详见 tasks/prd-nats-auction-core/k6-test-plan.md 第 4、6 节。

export const BASE = __ENV.BASE_URL || 'http://localhost:9000';
export const WS = __ENV.WS_URL || 'ws://localhost:9000';
export const TOKEN = __ENV.TOKEN || '';

// 由 setup 创建并 start 的竞拍;steady/stress 等场景复用同一场竞拍。
export const AUCTION_ID = __ENV.AUCTION_ID || '';

// 受理后等待 bid.placed 的超时;超时未见事件即计丢失。
export const EVENT_TIMEOUT_MS = Number(__ENV.EVENT_TIMEOUT_MS || 5000);
// 单 VU 出价间隔与会话时长。
export const BID_INTERVAL_MS = Number(__ENV.BID_INTERVAL_MS || 500);
export const SESSION_MS = Number(__ENV.SESSION_MS || 30000);

export function authHeaders(extra = {}) {
  const h = { 'Content-Type': 'application/json', ...extra };
  if (TOKEN) h['Authorization'] = `Bearer ${TOKEN}`;
  return h;
}

// SLO 闸门。数值为占位,待 PRD Open Questions 的 P95/吞吐量化目标确定后收紧。
export const thresholds = {
  'bid_accept_latency_ms': ['p(95)<200'],
  'bid_e2e_latency_ms': ['p(95)<1500', 'p(99)<3000'],
  'bid_events_lost': ['count<1'],
  'bid_roundtrip_ok': ['rate>0.999'],
  'ws_subscribed': ['rate>0.99'],
  'http_req_failed': ['rate<0.01'],
};
