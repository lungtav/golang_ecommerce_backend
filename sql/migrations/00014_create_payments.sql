-- +goose Up
create table if not exists payments(
    id uuid primary key default gen_random_uuid(),
    order_id uuid not null unique references orders(id),
    provider varchar(20) not null,
    provider_payment_id text not null,
    currency varchar(3) not null default 'NGN'
        check (currency in ('NGN')),
    amount integer not null,
    status varchar(10) not null default 'pending'
        check (status in ('succeeded', 'underpaid', 'failed')),
    payment_method varchar(20) not null,
    paid_at timestamptz not null default now(),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- +goose Down
drop table if exists payments;