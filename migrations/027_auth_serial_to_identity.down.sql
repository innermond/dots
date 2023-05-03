alter table auth drop constraint if exists auth_id_seq;
create sequence if not exists auth_id_seq as int4 start 1;
alter table auth drop id;
alter table auth add column id int4 default nextval('auth_id_seq') primary key;
