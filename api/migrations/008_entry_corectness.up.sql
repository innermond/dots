CREATE OR REPLACE FUNCTION entry_type_company_same_tid()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
declare
	company_tid ksuid;
	entry_type_tid ksuid;
	uid ksuid;
begin
	select id from "user" where id = NEW.tid into uid;
	if uid is null then
		raise exception 'not found uid';
	end if;

	if NEW.company_id is not null then
		select tid from company where id = NEW.company_id into company_tid;
		if company_tid != NEW.tid then
			raise exception 'company % has not expected tid %', company_tid, NEW.tid;
		end if;
	end if;

	if NEW.entry_type_id is not null then
		select tid from entry_type where id = NEW.entry_type_id into entry_type_tid;
		if entry_type_tid != NEW.tid then
			raise exception 'entry type % has not expected tid %', entry_type_tid, NEW.tid;
		end if;
	end if;

	if  company_tid != entry_type_tid then
		raise exception 'company % and entry type % has not the same expected tid %', company_tid, entry_type_tid, NEW.tid;
	end if;

	raise notice 'new tid: % old tid %', NEW.tid, OLD.tid;
	return NEW;
end;
$function$
;

drop trigger if exists entry_type_company_has_same_tid_tg on entry;
create trigger entry_type_company_has_same_tid_tg before insert or update on entry
for each row
execute function entry_type_company_same_tid();

alter table entry add constraint check_entry_quantity check (quantity > 0);
alter table entry alter column quantity drop default;
alter table entry alter column tid set not null;
