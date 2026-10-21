-- name: CreateStore :one
insert into stores (name, owner_id, description) values ($1, $2, $3) returning id, name, owner_id, description, created_at;

-- name: UpdateStore :execrows
update stores set name = $1, description = $2 where id =$3;

-- name: GetStore :one
select id, name, owner_id, description, created_at from stores where id = $1;