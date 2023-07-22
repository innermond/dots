create schema if not exists mock;

create table mock.word (v text);
copy mock.word (v) from '/tmp/words' where v not like '%''%';

create or replace function mock.generate_random_title(a integer default 1, z integer default 4)
 returns text
 language sql
as $function$
  select initcap(array_to_string(array(
    select v from mock.word order by random() limit (select floor(random()*(z-a+1)+a)::int)
  ), ' '))
$function$
;

create or replace function mock.gen_rnd_tid() returns text as $$
	select id from (values ('2PH25DxmohuFCf3w73fQSTLJeVO'), ('2PH24UhBlN5tlYdAmpdwiyPuWgB')) a(id) order by random() limit 1;
$$ language sql;

create or replace function mock.create_company() returns void  as $$
	insert into api.company
	(longname, tin, rn, tid)
	values 
	(
		mock.generate_random_title(3, 4),
		substr(md5(random()::text), 1, 12), 
		substr(md5(random()::text), 1, 8), 
		mock.gen_rnd_tid()
	)
$$ language sql;
