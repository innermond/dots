create or replace function mock.generate_random_title(a integer default 1, z integer default 4)
 returns text
 language sql
as $function$
  select initcap(array_to_string(array(
    select v from mock.word order by random() limit (select floor(random()*(z-a+1)+a)::int)
  ), ' '))
$function$
;

create or replace function mock.gen_rnd_unit()
 returns text
 language sql
as $function$
	select id from (values ('unit'), ('buc'), ('piece'), ('cm'), ('mm'), ('pcs'), ('sq'), ('square meter'), ('piese'), ('hour'), ('rola')) a(id) order by random() limit 1;
$function$
;

create or replace function mock.create_entry_type()
 returns void
 language sql
as $function$
	insert into api.entry_type
	(code, description, unit, tid)
	values 
	(
		concat(substr(md5(random()::text), 1, 6),'-', substr(md5(random()::text), 1, 4)),
		mock.generate_random_title(3, 10),
		mock.gen_rnd_unit(),
		mock.gen_rnd_tid()
	)
$function$
;

create or replace function mock.create_entry(dispersed boolean default true, entry_type_id integer default null, company_id integer default null, a integer default 1, z integer default 1000)
 returns void
 language plpgsql
as $function$
declare vtid api.ksuid;
declare etid integer default null;
declare cid integer default null;
begin
	select mock.gen_rnd_tid() into vtid;
	if dispersed then
		insert into api.entry
		(entry_type_id, quantity, company_id, tid)
		values 
		(
			(select id from api.entry_type et where et.tid = vtid order by random() limit 1),
			(select floor(random()*(z-a+1))+a),
			(select id from api.company c where c.tid = vtid order by random() limit 1),
			vtid
		);
	else
		select id, et.tid into etid, vtid from api.entry_type et where id = entry_type_id;
		select id into cid from api.company c where id = company_id and c.tid = vtid;
		if etid is null or cid is null then
			raise exception 'entry type %', entry_type_id;
		end if;
		insert into api.entry
		(entry_type_id, quantity, company_id, tid)
		values 
		(
			etid,
			(select floor(random()*(z-a+1))+a),
			cid,
			vtid
		);		
	
	end if;
end;
$function$
;
