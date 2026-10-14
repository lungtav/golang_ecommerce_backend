-- name: CreateVerficationToken :exec
insert into verification_tokens (token_hash, user_id) values($1, $2);