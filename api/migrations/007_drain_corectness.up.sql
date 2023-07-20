create or replace function api.deed_entry_same_tid() returns trigger as $$
declare 
	has_not_same boolean;
	deed_tid api.ksuid;
	entry_tid api.ksuid;
begin
	select tid from api.deed where id = NEW.deed_id into deed_tid;
	select tid from api.entry where id = NEW.entry_id into entry_tid;
	has_not_same = (
		(deed_tid is not null and entry_tid is not null) and 
		(deed_tid != entry_tid or (deed_tid != NEW.tid or entry_tid != NEW.tid))
	);
	if  has_not_same then 
		raise exception 'deed % and entry % has not the same expected tid %', deed_tid, entry_tid, NEW.tid; 	
	end if;
	return NEW;
end;
$$ language plpgsql;

drop trigger if exists deed_entry_has_same_tid_tg on api.drain;
create trigger deed_entry_has_same_tid_tg before insert or update on api.drain 
for each row
execute function api.deed_entry_same_tid();
