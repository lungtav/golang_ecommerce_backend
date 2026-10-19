-- name: CreateStore :one
insert into stores (name, owner_id, description) values ($1, $2, $3) returning id, name, owner_id, description, created_at;