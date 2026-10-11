-- +goose Up
create table if not exists product_variants(
    id uuid primary key default gen_random_uuid(),
    product_id uuid not null references products(id),
    sku text not null,
    size varchar(5),
    color varchar(10),
    price_in_kobo integer not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists products_id_variants_idx on product_variants(product_id);

-- +goose Down
drop table if exists product_variants;