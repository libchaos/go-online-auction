-- name: CreateWithdrawal :one
INSERT INTO withdrawals (
    user_id, ledger_account_id, alipay_account, alipay_real_name, amount_cents, currency, status, out_biz_no, frozen_op_id, version, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
)
RETURNING id, user_id, ledger_account_id, alipay_account, alipay_real_name, amount_cents, currency, status, out_biz_no, frozen_op_id, alipay_order_id, fail_reason, version, created_at, updated_at;

-- name: GetWithdrawalByID :one
SELECT id, user_id, ledger_account_id, alipay_account, alipay_real_name, amount_cents, currency, status, out_biz_no, frozen_op_id, alipay_order_id, fail_reason, version, created_at, updated_at
FROM withdrawals
WHERE id = $1;

-- name: GetWithdrawalByOutBizNo :one
SELECT id, user_id, ledger_account_id, alipay_account, alipay_real_name, amount_cents, currency, status, out_biz_no, frozen_op_id, alipay_order_id, fail_reason, version, created_at, updated_at
FROM withdrawals
WHERE out_biz_no = $1;

-- name: UpdateWithdrawal :one
UPDATE withdrawals
SET status = $1, alipay_order_id = $2, fail_reason = $3, version = $4, updated_at = NOW()
WHERE id = $5 AND version = $6
RETURNING id, user_id, ledger_account_id, alipay_account, alipay_real_name, amount_cents, currency, status, out_biz_no, frozen_op_id, alipay_order_id, fail_reason, version, created_at, updated_at;
