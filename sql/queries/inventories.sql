-- name: CreateInventory :exec
insert into inventories (product_variant_id, store_id, quantity) values ($1, $2, $3);