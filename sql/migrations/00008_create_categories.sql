-- +goose Up
create table if not exists categories(
    id uuid primary key default gen_random_uuid(),
    name varchar(100) not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- +goose Down
drop table if exists categories;