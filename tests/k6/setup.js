// 测试数据准备 helper:登录 → 创建 auction → start,返回 { token, auctionID }。
// 供各场景的 setup() 复用;也可单独运行导出 AUCTION_ID。
// 注意:端点路径/请求体需按实际 API(auction_handler.go / dto)核对。
import http from 'k6/http';
import { check, fail } from 'k6';
import { BASE } from './config.js';

export function login(email, password) {
  const res = http.post(`${BASE}/api/v0/auth/login`,
    JSON.stringify({ email, password }),
    { headers: { 'Content-Type': 'application/json' } });
  check(res, { 'login 200': (r) => r.status === 200 }) || fail(`login failed: ${res.status} ${res.body}`);
  return JSON.parse(res.body).access_token; // 按实际响应字段名调整
}

export function createAndStartAuction(token) {
  const headers = { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` };

  const create = http.post(`${BASE}/api/v0/auctions`,
    JSON.stringify({ /* listing_id, start_time, end_time, ... 按实际 DTO 填 */ }),
    { headers });
  check(create, { 'create 201': (r) => r.status === 201 }) || fail(`create failed: ${create.status} ${create.body}`);
  const auctionID = JSON.parse(create.body).id;

  const start = http.post(`${BASE}/api/v0/auctions/${auctionID}/start`, null, { headers });
  check(start, { 'start 2xx': (r) => r.status >= 200 && r.status < 300 }) || fail(`start failed: ${start.status}`);

  return auctionID;
}

// 便捷组合:一次性拿到 token + 已开始的 auctionID。
export function prepare() {
  const token = login(__ENV.TEST_EMAIL || 'load@test.local', __ENV.TEST_PASSWORD || 'password');
  const auctionID = __ENV.AUCTION_ID || createAndStartAuction(token);
  return { token, auctionID };
}
