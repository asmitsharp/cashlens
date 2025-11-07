-- name: GetAllGlobalRules :many
SELECT * FROM global_categorization_rules
WHERE is_active = TRUE
ORDER BY priority DESC, keyword ASC;

-- name: GetGlobalRuleByKeyword :one
SELECT * FROM global_categorization_rules
WHERE keyword = $1 AND is_active = TRUE
LIMIT 1;

-- name: CreateGlobalRule :one
INSERT INTO global_categorization_rules (keyword, category, priority, match_type, similarity_threshold, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateGlobalRule :one
UPDATE global_categorization_rules
SET category = $2, priority = $3, match_type = $4, similarity_threshold = $5, is_active = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGlobalRule :exec
DELETE FROM global_categorization_rules
WHERE id = $1;

-- name: DeactivateGlobalRule :exec
UPDATE global_categorization_rules
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: GetUserRules :many
SELECT * FROM user_categorization_rules
WHERE user_id = $1 AND is_active = TRUE
ORDER BY priority DESC, keyword ASC;

-- name: GetUserRuleByKeyword :one
SELECT * FROM user_categorization_rules
WHERE user_id = $1 AND keyword = $2 AND is_active = TRUE
LIMIT 1;

-- name: CreateUserRule :one
INSERT INTO user_categorization_rules (user_id, keyword, category, priority, match_type, similarity_threshold, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateUserRule :one
UPDATE user_categorization_rules
SET category = $2, priority = $3, match_type = $4, similarity_threshold = $5, is_active = $6, updated_at = NOW()
WHERE id = $1 AND user_id = $7
RETURNING *;

-- name: DeleteUserRule :exec
DELETE FROM user_categorization_rules
WHERE id = $1 AND user_id = $2;

-- name: DeactivateUserRule :exec
UPDATE user_categorization_rules
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: GetAllRulesForUser :many
-- Get combined list of global and user rules, user rules take precedence
SELECT
    u.id,
    u.keyword,
    u.category,
    u.priority,
    'user' as rule_type
FROM user_categorization_rules u
WHERE u.user_id = $1 AND u.is_active = TRUE

UNION ALL

SELECT
    g.id,
    g.keyword,
    g.category,
    g.priority,
    'global' as rule_type
FROM global_categorization_rules g
WHERE g.is_active = TRUE
AND g.keyword NOT IN (
    SELECT u2.keyword FROM user_categorization_rules u2
    WHERE u2.user_id = $1 AND u2.is_active = TRUE
)

ORDER BY priority DESC, keyword ASC;

-- name: GetCategoriesWithRuleCount :many
SELECT
    category,
    COUNT(*) as rule_count
FROM (
    SELECT category FROM global_categorization_rules WHERE is_active = TRUE
    UNION ALL
    SELECT category FROM user_categorization_rules WHERE user_id = $1 AND is_active = TRUE
) AS all_rules
GROUP BY category
ORDER BY rule_count DESC;

-- name: SearchRulesByKeyword :many
SELECT * FROM global_categorization_rules
WHERE keyword ILIKE '%' || $1 || '%' AND is_active = TRUE
ORDER BY priority DESC, keyword ASC
LIMIT $2;

-- name: SearchUserRulesByKeyword :many
SELECT * FROM user_categorization_rules
WHERE user_id = $1 AND keyword ILIKE '%' || $2 || '%' AND is_active = TRUE
ORDER BY priority DESC, keyword ASC
LIMIT $3;

-- name: GetRuleStats :one
SELECT
    (SELECT COUNT(*) FROM global_categorization_rules g WHERE g.is_active = TRUE) as global_rules_count,
    (SELECT COUNT(*) FROM user_categorization_rules u WHERE u.user_id = $1 AND u.is_active = TRUE) as user_rules_count,
    (SELECT COUNT(DISTINCT g2.category) FROM global_categorization_rules g2 WHERE g2.is_active = TRUE) as global_categories_count,
    (SELECT COUNT(DISTINCT u2.category) FROM user_categorization_rules u2 WHERE u2.user_id = $1 AND u2.is_active = TRUE) as user_categories_count;
