-- +goose Up
create table if not exists orders(
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users(id),
    status varchar(10) not null default 'pending'
        check (status in ('successful', 'pending', 'failed')),
    amount integer not null,
    currency varchar(3) not null default 'NGN'
        check (currency in ('NGN')),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists user_order_idx on orders(user_id);

-- +goose Down
drop table if exists orders;