-- +goose Up
create table if not exists verification_tokens(
    id uuid primary key default gen_random_uuid(),
    token_hash varchar(255) not null unique,
    created_at timestamptz not null default now(),
    used_at timestamptz,
    expires_at timestamptz not null default now() + interval '15 mins',
    user_id uuid not null references users(id)
);


-- +goose Down
drop table if exists verification_tokens;