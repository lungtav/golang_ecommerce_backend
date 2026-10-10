-- +goose Up
create table if not exists products(
    id uuid primary key default gen_random_uuid(),
    store_id uuid not null references stores(id),
    name varchar(100) not null,
    status varchar(100) not null default 'draft'
        check (status in ('archived', 'active', 'draft')),
    description text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists store_id_products_idx on products(store_id);

-- +goose Down
drop table if exists products;