-- +goose Up
create table if not exists carts(
    id uuid primary key default gen_random_uuid(),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    user_id uuid not null unique references users(id)
);

-- +goose Down
drop table if exists carts;