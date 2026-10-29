-- name: CreateCategory :exec
insert into categories (name) values ($1);