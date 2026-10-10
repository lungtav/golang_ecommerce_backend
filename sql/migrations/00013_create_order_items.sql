-- +goose Up
create table if not exists order_items(
    product_id uuid not null references products(id),
    product_variant_id uuid not null references products_variant(id),
    order_id uuid not null references orders(id),
    quantity integer not null default '1'
        check (quantity >= 1),
    created_at timestamptz not null default now()
);

create index if not exists product_variant_store_idx on inventory(product_variant_id, order_id);

-- +goose Down
drop table if exists order_items
