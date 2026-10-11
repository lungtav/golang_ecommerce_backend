-- +goose Up
create table if not exists product_categories(
    product_id uuid not null references products(id),
    category_id uuid not null references categories(id),
    created_at timestamptz not null default now()
);

create index if not exists product_categories_idx on product_categories(product_id, category_id);

-- +goose Down
drop table if exists product_categories;