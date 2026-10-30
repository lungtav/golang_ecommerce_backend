-- name: CreateInventory :exec
insert into inventories (product_variant_id, store_id, quantity) values ($1, $2, $3);

-- name: GetInventory :one
select quantity from inventories  where product_variant_id = $1;

-- name: DecrementInventory :execrows
update inventories set quantity = quantity - $2, updated_at = now() where product_variant_id = $1 and quantity >= $2;

-- name: ClearCart :exec
delete from cart_items where cart_id = $1;