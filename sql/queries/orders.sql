-- name: CreateOrder :one
insert into orders (user_id, status, amount, currency) values ($1 , $2, $3, $4) returning id;

-- name: CreateOrderItems :exec
insert into order_items (product_id, product_variant_id, order_id, quantity) values ($1, $2, $3, $4);