-- name: CreatePayment :one
insert into payments (order_id, amount_in_kobo) values ($1, $2) returning id;

-- name: MarkPaymentFailed :exec
update payments set status = 'failed', updated_at = now() where id = $1 and status = 'pending';

-- name: MarkPaymentSucceeded :execrows
UPDATE payments
SET status = 'succeeded',
    paid_at = now(),
    payment_method = $2,
    provider_payment_id = $3,
    updated_at = now()
WHERE id = $1 AND status = 'pending';

-- name: MarkOrderPaid :execrows
UPDATE orders
SET status = 'paid', updated_at = now()
WHERE id = $1 AND status = 'pending';

-- name: GetPaymentByID :one
SELECT id, order_id, amount_in_kobo, status
FROM payments
WHERE id = $1;