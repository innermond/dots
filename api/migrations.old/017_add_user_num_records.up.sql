create table package (
  name text not null unique check (name in ('one eye', 'two eyes', 'three eyes')),
  company int2 not null,
  deed int4 not null,
  drain int8 not null,
  entry_type int4 not null,
  entry int8 not null
);

insert into package values
('one eye', 1, 10, 30, 3, 9), -- just to visit and play
('two eyes', 2, 20, 60, 6, 18),
('three eyes', 3, 30, 90, 9, 27);

alter table "user" add column package_kind text not null default 'one eye' check (package_kind in ('one eye', 'two eyes', 'three eyes'));

-- personalized limits per user
-- allows nulls as null value means it has value from package
create table user_num_record (
  user_id int4 not null unique references "user" (id),
  company int2 null,
  deed int4 null,
  drain int8 null,
  entry_type int4 null,
  entry int8 null
);

insert into user_num_record values
(2, 10, null, null, null, 1);
