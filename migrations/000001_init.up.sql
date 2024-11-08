create extension if not exists "uuid-ossp";

create table if not exists "conversations" (
  id uuid primary key default uuid_generate_v4(),
  name varchar,
  owner_id uuid not null,
  created_at timestamp default now(),
  updated_at timestamp
);

create table if not exists "messages" (
  id serial primary key,
  conversation_id uuid not null,
  is_user boolean not null,
  body text not null,
  created_at timestamp default now(),
  updated_at timestamp,
  foreign key (conversation_id) references conversations(id)
);