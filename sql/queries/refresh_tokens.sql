-- name: CreateRefreshToken :one
insert into refresh_tokens (token_hash, user_id) values ($1, $2) returning *;

-- name: GetRefreshToken :one
select token_hash, revoked_at, expires_at, user_id from refresh_tokens where token_hash =$1;

