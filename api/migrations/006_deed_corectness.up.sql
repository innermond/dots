create or replace function company_has_same_tid() returns trigger as $$
declare has_same boolean;
begin
	select exists(select id from company c where c.id=NEW.company_id and c.tid=NEW.tid) into has_same;
	if not has_same then
		raise exception 'company % has not the same tid % or not exists', NEW.company_id, NEW.tid;
	end if;
	return NEW;
end;
$$ language plpgsql;

drop trigger if exists company_has_same_tid_tg on deed;
create trigger company_has_same_tid_tg before insert or update on deed
for each row
execute function company_has_same_tid(); 
