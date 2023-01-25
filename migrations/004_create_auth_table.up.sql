create table "auth" (
  id serial primary key,
  user_id integer not null references "user" (id) on delete cascade,
  source text not null,
  source_id text not null,
  access_token text not null,
  refresh_token text not null,
  expiry timestamp,
  created_at timestamp not null,
  updated_at timestamp not null
);
