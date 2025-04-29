-- name: SaveAuditLog :exec
INSERT INTO audit_logs(
    user_id, affected_account_id, action, metadata
) VALUES ($1, $2, $3, $4);

-- name: GetAccountStatusHistory :many
SELECT
    a.id AS account_id,
    a.status AS current_status,
    al.action AS action,
    COALESCE(al.metadata->>'old_status', '')::varchar AS old_status,
    COALESCE(al.metadata->>'new_status', '')::varchar AS new_status,
    u.first_name || ' ' || u.last_name AS action_by,
    al.created_at AS created_at
FROM
    audit_logs al
    JOIN
        accounts a ON al.affected_account_id = a.id
    JOIN
        users u ON al.user_id = u.id
WHERE
    al.affected_account_id = $1
  AND al.action = 'account_status_change'
ORDER BY
    al.created_at DESC;

-- name: GetAuditLogsForAccount :many
SELECT
    a.id AS account_id,
    a.status AS current_status,
    al.action AS action_code,
    CASE al.action
        WHEN 'create_account' THEN 'Created Account'
        WHEN 'account_credit' THEN 'Credited Account'
        WHEN 'account_debit' THEN 'Debited Account'
        WHEN 'account_status_change' THEN COALESCE(al.metadata->>'new_status', '') || ' Account'
        ELSE al.action -- Keep the original action if not one of the defined ones
        END AS action,
    COALESCE(al.metadata->>'old_status', '')::varchar AS old_status,
    COALESCE(al.metadata->>'new_status', '')::varchar AS new_status,
    COALESCE(al.metadata->>'currency', '')::varchar AS currency,
    CAST(COALESCE(al.metadata->>'amount', '0') AS BIGINT) AS amount,
    u.first_name || ' ' || u.last_name AS action_by,
    al.created_at AS created_at
FROM
    audit_logs al
    JOIN
    accounts a ON al.affected_account_id = a.id
    JOIN
    users u ON al.user_id = u.id
WHERE
    al.affected_account_id = 'dcb947af-84cd-4f53-a573-9bff39ba9eec'
ORDER BY
    al.created_at DESC;