alter table auth drop constraint if exists auth_id_pkey;
alter table auth drop id;
alter table auth add column id int4 generated always as identity primary key;
