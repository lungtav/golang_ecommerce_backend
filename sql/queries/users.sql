-- name: GetUserByEmail :one
select id, name, email, email_verified_at from users where email = $1 limit 1;

-- name: CreateUser :one
insert into users (email, password_hash, name) values ($1, $2, $3) returning *;

-- name: VerifyUser :exec
update users set email_verified_at = $1 where email =$2;