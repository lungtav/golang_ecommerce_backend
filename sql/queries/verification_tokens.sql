-- name: CreateVerificationToken :exec
insert into verification_tokens (token_hash, user_id) values($1, $2);

-- name: GetVerificationToken :one
select token_hash, used_at, created_at, expires_at from verification_tokens where user_id = $1;

-- name: DeleteVerificationToken :exec
delete from verification_tokens where user_id = $1;

-- name: GetVerificationTokenByHash :one
select id, user_id, used_at, expires_at, created_at from verification_tokens where token_hash = $1;

-- name: UpdateVerificationTokensById :execrows
update verification_tokens set used_at = now() where id =$1 and used_at is null;