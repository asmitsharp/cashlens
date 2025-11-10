-- name: GetKPIs :one
SELECT
    COALESCE(SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE 0 END), 0) AS total_inflow,
    COALESCE(SUM(CASE WHEN txn_type = 'debit' THEN amount ELSE 0 END), 0) AS total_outflow,
    COALESCE(SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE -amount END), 0) AS net_cash_flow,
    COUNT(*) AS transaction_count
FROM transactions
WHERE user_id = $1
  AND txn_date BETWEEN $2 AND $3;

-- name: GetNetFlowTrend :many
SELECT
    DATE_TRUNC(sqlc.arg(date_trunc)::text, txn_date::timestamp)::timestamp AS period,
    COALESCE(SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE -amount END), 0) AS net_flow
FROM transactions
WHERE user_id = $1
  AND txn_date BETWEEN $2 AND $3
GROUP BY DATE_TRUNC(sqlc.arg(date_trunc)::text, txn_date::timestamp)
ORDER BY period;

-- name: GetCashFlowTrend :many
SELECT
    DATE_TRUNC(sqlc.arg(date_trunc)::text, txn_date::timestamp)::timestamp AS period,
    COALESCE(SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE 0 END), 0) AS inflow,
    COALESCE(SUM(CASE WHEN txn_type = 'debit' THEN amount ELSE 0 END), 0) AS outflow,
    COALESCE(SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE -amount END), 0) AS net_flow
FROM transactions
WHERE user_id = $1
  AND txn_date BETWEEN $2 AND $3
GROUP BY DATE_TRUNC(sqlc.arg(date_trunc)::text, txn_date::timestamp)
ORDER BY period;
