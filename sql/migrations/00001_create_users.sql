-- +goose Up
create table if not exists users(
    id uuid primary key default gen_random_uuid(),
    name varchar(100) not null,
    email varchar(255) unique not null,
    email_verified_at timestamptz,
    password_hash varchar(255) not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- +goose Down
drop table if exists users;