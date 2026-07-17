// 自定义指标 — 供各 k6 场景脚本共享。
// 详见 tasks/prd-nats-auction-core/k6-test-plan.md 第 2 节。
import { Trend, Counter, Rate } from 'k6/metrics';

// 受理(HTTP 202) → 收到对应 bid.placed(WS) 的端到端延迟(ms)。
// 新架构最关键的新指标:衡量异步管道(命令流→落库→事件流→WS)整体延迟。
export const bidE2ELatency = new Trend('bid_e2e_latency_ms', true);

// HTTP 出价受理延迟(仅接受层)。应与 DB/落库解耦而保持很低。
export const bidAcceptLatency = new Trend('bid_accept_latency_ms', true);

// 已被 202 受理、但在超时窗口内未收到对应 WS 事件 → 计为丢失。
export const bidEventsLost = new Counter('bid_events_lost');

// 成功闭环(受理且在超时内收到事件)的比率。
export const bidRoundtripOk = new Rate('bid_roundtrip_ok');

// WS 订阅确认(subscription_confirmed)成功率。
export const wsSubscribed = new Rate('ws_subscribed');
