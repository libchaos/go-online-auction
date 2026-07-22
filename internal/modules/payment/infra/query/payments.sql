-- name: CreatePayment :one
INSERT INTO payments (
    user_id, amount_cents, currency, status, out_trade_no, qr_code_url, version, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
)
RETURNING id, user_id, amount_cents, currency, status, out_trade_no, qr_code_url, alipay_trade_no, version, created_at, updated_at;

-- name: GetPaymentByID :one
SELECT id, user_id, amount_cents, currency, status, out_trade_no, qr_code_url, alipay_trade_no, version, created_at, updated_at
FROM payments
WHERE id = $1;

-- name: GetPaymentByOutTradeNo :one
SELECT id, user_id, amount_cents, currency, status, out_trade_no, qr_code_url, alipay_trade_no, version, created_at, updated_at
FROM payments
WHERE out_trade_no = $1;

-- name: UpdatePayment :one
UPDATE payments
SET status = $1, alipay_trade_no = $2, version = $3, updated_at = NOW()
WHERE id = $4 AND version = $5
RETURNING id, user_id, amount_cents, currency, status, out_trade_no, qr_code_url, alipay_trade_no, version, created_at, updated_at;
