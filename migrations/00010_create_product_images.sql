-- +goose Up
create table if not exists product_images(
    id uuid primary key default gen_random_uuid(),
    product_variant_id uuid not null unique references product_variants(id),
    image_url text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- +goose Down
drop table if exists product_images;