-- +goose Up
create table teams (
   name text primary key
);

create table users (
   id text primary key,
   username text not null,
   team_name text not null references teams(name),
   is_active boolean not null default true,
   created_at timestamptz not null default now(),
   updated_at timestamptz not null default now()
);

create table pull_requests (
   id text primary key,
   name text not null,
   author_id text not null references users(id),
   status text not null check (status in ('OPEN', 'MERGED')),
   created_at timestamptz not null default now(),
   merged_at timestamptz
);

-- Таблица ревьюверов для PR
create table pr_reviewers (
      pull_request_id text not null references pull_requests(id),
      reviewer_id text references users(id),
      unique (pull_request_id, reviewer_id)
);