-- name: CreateRefreshToken :exec
insert into refresh_tokens (token_hash, user_id) values ($1, $2);

-- name: GetRefreshToken :one
select id, revoked_at, expires_at, user_id from refresh_tokens where token_hash =$1;

-- name: RevokeRefreshToken :execrows
update refresh_tokens set revoked_at = now() where id = $1 and revoked_at is null;