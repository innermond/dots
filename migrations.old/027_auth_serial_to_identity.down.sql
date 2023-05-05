alter table auth alter column id drop identity if exists;
alter table auth alter column id drop default;
drop sequence if exists auth_id_seq;

do $$
declare max_id int4;
begin
select max(id)+1 into max_id from auth;
execute 'create sequence if not exists auth_id_seq as int4 start ' || max_id || ' owned by auth.id';
alter table auth alter column id drop identity if exists;
alter table auth alter column id set default nextval('auth_id_seq');
end$$;
