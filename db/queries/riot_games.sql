-- name: GetRiotGamesByMatchIDs :many
SELECT id, match_id, played_at, won, analysis, created_at, updated_at
FROM riot_games
WHERE match_id = ANY(sqlc.narg('match_ids')::text[]);

-- name: GetRiotGamesInRange :many
SELECT id, match_id, played_at, won, analysis, created_at, updated_at
FROM riot_games
WHERE played_at >= $1 AND played_at < $2
ORDER BY played_at DESC;