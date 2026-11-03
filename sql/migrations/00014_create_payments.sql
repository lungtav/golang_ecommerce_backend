-- +goose Up
create table if not exists payments(
    id uuid primary key default gen_random_uuid(),
    order_id uuid not null references orders(id),
    provider varchar(20) not null default 'paystack',
    provider_payment_id text,
    currency varchar(3) not null default 'NGN'
        check (currency in ('NGN')),
    amount_in_kobo bigint not null check (amount_in_kobo > 0),
    status varchar(10) not null default 'pending'
        check (status in ('succeeded', 'underpaid', 'failed', 'pending')),
    payment_method varchar(20),
    paid_at timestamptz default now(),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists payments_order_id_idx on payments(order_id);

-- +goose Down
drop table if exists payments;