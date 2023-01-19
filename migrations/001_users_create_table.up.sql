create table if not exists users (
  id serial primary key,
  name varchar(50) not null,
  created_on timestamp not null,
  last_login timestamp
);
