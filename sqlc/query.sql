-- Categories

-- name: CountCategories :one
SELECT COUNT(*) FROM categories WHERE deleted_at IS NULL;

-- name: ListCategoriesPaginated :many
SELECT id, slug, created_at, updated_at, deleted_at
FROM categories
WHERE deleted_at IS NULL
ORDER BY updated_at
LIMIT $1 OFFSET $2;

-- name: CreateCategory :one
INSERT INTO categories (slug)
VALUES ($1)
RETURNING *;

-- name: GetCategory :one
SELECT * FROM categories
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListCategories :many
SELECT * FROM categories
WHERE deleted_at IS NULL
ORDER BY updated_at;

-- name: UpdateCategory :one
UPDATE categories
SET slug = $2,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteCategory :exec
UPDATE categories
SET deleted_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- Series

-- name: CountSeries :one
SELECT COUNT(*) FROM series WHERE deleted_at IS NULL;

-- name: ListSeriesPaginated :many
SELECT id, title, description, category_id, language, series_type,
       created_at, updated_at, deleted_at 
FROM series
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateSeries :one
INSERT INTO series (title, description, category_id, language, series_type)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSeries :one
SELECT * FROM series
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListSeries :many
SELECT * FROM series
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateSeries :one
UPDATE series
SET title = $2,
    description = $3,
    category_id = $4,
    language = $5,
    series_type = $6,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteSeries :exec
UPDATE series
SET deleted_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- Episodes

-- name: CountEpisodesBySeries :one
SELECT COUNT(*) FROM episodes
WHERE series_id = $1 AND deleted_at IS NULL;

-- name: ListEpisodesBySeriesPaginated :many
SELECT id, series_id, title, description, duration_seconds,
       publish_date, created_at, updated_at, deleted_at
FROM episodes
WHERE series_id = $1 AND deleted_at IS NULL
ORDER BY publish_date DESC
LIMIT $2 OFFSET $3;

-- name: CreateEpisode :one
INSERT INTO episodes (
    series_id, title, description,
    duration_seconds, publish_date
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetEpisode :one
SELECT * FROM episodes
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListEpisodesBySeries :many
SELECT * FROM episodes
WHERE series_id = $1
  AND deleted_at IS NULL
ORDER BY publish_date DESC;

-- name: UpdateEpisode :one
UPDATE episodes
SET title = $2,
    description = $3,
    duration_seconds = $4,
    publish_date = $5,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteEpisode :exec
UPDATE episodes
SET deleted_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- Episode Assets

-- name: ListAssetsByEpisode :many
SELECT * FROM episode_assets
WHERE episode_id = $1;

-- name: CreateAsset :one
INSERT INTO episode_assets (
    episode_id, asset_type, mime_type, size_bytes, url, storage
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAsset :one
SELECT * FROM episode_assets
WHERE id = $1;

-- name: UpdateAsset :one
UPDATE episode_assets
SET mime_type = $2,
    size_bytes = $3,
    url = $4,
    storage = $5
WHERE id = $1
RETURNING *;

-- name: DeleteAsset :exec
DELETE FROM episode_assets
WHERE id = $1;

-- name: GetEpisodeWithAssets :many
SELECT 
    e.id               AS episode_id,
    e.series_id,
    e.title,
    e.description,
    e.duration_seconds,
    e.publish_date,
    e.created_at       AS episode_created_at,
    e.updated_at       AS episode_updated_at,
    a.id               AS asset_id,
    a.asset_type,
    a.mime_type,
    a.size_bytes,
    a.url,
    a.storage,
    a.created_at       AS asset_created_at
FROM episodes e
LEFT JOIN episode_assets a ON e.id = a.episode_id
WHERE e.id = $1 AND e.deleted_at IS NULL;

-- name: ListEpisodesWithAssetsBySeriesPaginated :many
SELECT
    e.id AS episode_id,
    e.series_id,
    e.title,
    e.description,
    e.duration_seconds,
    e.publish_date,
    e.created_at AS episode_created_at,
    e.updated_at AS episode_updated_at,
    a.id AS asset_id,
    a.asset_type,
    a.mime_type,
    a.size_bytes,
    a.url,
    a.storage,
    a.created_at AS asset_created_at
FROM
    episodes e
LEFT JOIN
    episode_assets a ON e.id = a.episode_id
WHERE
    e.series_id = $1 AND e.deleted_at IS NULL
ORDER BY
    e.publish_date DESC
LIMIT $2 OFFSET $3;