-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT feeds.name AS feed_name, feeds.url AS feed_url, users.name AS user_name
FROM feeds
INNER JOIN users ON feeds.user_id = users.id;

-- name: GetFeedIdByURL :one
SELECT id
FROM feeds
WHERE url = $1;

-- name: CreateFeedFollow :one
WITH feed_record AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT feed_record.id, feed_record.created_at, feed_record.updated_at, feed_record.user_id, feed_record.feed_id, users.name AS user_name, feeds.name AS feed_name
FROM feed_record
INNER JOIN users ON feed_record.user_id = users.id
INNER JOIN feeds ON feed_record.feed_id = feeds.id;

-- name: GetFeedFollowsForUser :many
SELECT feeds.id AS feed_id, feeds.name AS feed_name, users.name AS user_name
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
INNER JOIN users ON feed_follows.user_id = users.id
WHERE feed_follows.user_id = $1;

-- name: UnfollowFeed :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1,
    updated_at = $1
WHERE id = $2;

-- name: GetNextFeedToFetch :one
SELECT id, url
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;

-- name: CreatePost :exec
INSERT INTO posts (
    id,
    created_at,
    updated_at,
    title,
    url,
    description,
    published_at,
    feed_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetPostsForUser :many
SELECT title, published_at, url, description
FROM posts
WHERE feed_id = $1
ORDER BY published_at DESC
LIMIT $2;
