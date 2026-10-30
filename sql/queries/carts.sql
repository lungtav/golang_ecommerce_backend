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

-- name: GetCartItems :many
select ci.product_id, ci.product_variant_id, ci.quantity, p.name as product_name, pv.size, pv.price_in_kobo, (ci.quantity * pv.price_in_kobo) as line_total
from cart_items ci join carts c on c.id = ci.cart_id join products p on p.id = ci.product_id join product_variants pv on pv.id = ci.product_variant_id where c.user_id = $1 order by ci.created_at;