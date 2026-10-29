-- name: CreateCart :one
insert into carts (user_id) values($1) returning id;

-- name: AddToCart :exec
insert into cart_items (product_id, product_variant_id, cart_id, quantity) values ($1, $2, $3, $4);

-- name: GetCart :one
select id from carts where user_id = $1;

-- name: GetCartItemQuantity :one
select quantity from cart_items where cart_id = $1 and product_variant_id = $2;

-- name: UpdateCartItem :execrows
update cart_items set quantity = $1, updated_at = now() where cart_id = $2 and product_variant_id = $3;