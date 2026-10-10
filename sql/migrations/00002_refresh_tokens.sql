-- +goose Up
create table if not exists refresh_tokens(
    id uuid primary key gen_random_uuid(),
    token_hash varchar(255) not null unique,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    revoked_at timestamptz,
    expires_at timestamptz not null default now() + interval `15 days`,
    user_id uuid not null references users(id)
);


-- +goose Down
drop table if exists refresh_tokens;