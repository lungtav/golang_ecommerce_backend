-- name: GetUserByEmail :one
select id, name, email, password_hash, email_verified_at from users where email = $1 limit 1;

-- name: UpdateUserById :exec
update users set email_verified_at = now() where id =$1;

-- name: CreateUser :one
insert into users (email, password_hash, name) values ($1, $2, $3) returning *;

-- name: VerifyUser :exec
update users set email_verified_at = $1 where email =$2;