-- +goose Up
create table if not exists cart_items(
    product_id uuid not null references products(id),
    product_variant_id uuid not null references products_variants(id),
    cart_id uuid not null references carts(id),
    quantity integer not null default 1
        check (quantity >= 1),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);


create index if not exists variant_cart_idx on cart_items(product_variant_id, cart_id);
create index if not exists scart_id_items_idx on cart_items(cart_id);

-- +goose Down
drop table if exists cart_items;