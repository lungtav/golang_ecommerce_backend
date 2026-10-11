-- +goose Up
create table if not exists inventories(
    product_variant_id uuid not null references product_variants(id),
    store_id uuid not null references stores(id),
    quantity integer not null default 0
        check (quantity >= 0),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists product_variant_store_idx on inventories(product_variant_id, store_id);

-- +goose Down
drop table if exists inventories;