-- +goose Up
create table if not exists stores(
    id uuid primary key default gen_random_uuid(),
    name varchar(100) not null unique,
    owner_id uuid not null references users(id),
    description text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- +goose Down
drop table if exists stores;