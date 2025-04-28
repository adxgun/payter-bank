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