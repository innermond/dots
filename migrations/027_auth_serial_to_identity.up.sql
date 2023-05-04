alter table auth alter column id drop identity if exists;
alter table auth alter column id drop default;
alter table auth alter column id add generated always as identity;
