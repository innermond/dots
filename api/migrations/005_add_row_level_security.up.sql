create or replace function get_tenent() returns ksuid as $$
begin
	return nullif(current_setting('app.uid', true), '');
end;
$$ language plpgsql;

alter table entry add column tid ksuid;
update entry e set tid = c.tid from company c where c.id = e.company_id;
alter table entry alter tid set default not null;
alter table entry add constraint entry_tid_fk_user_id foreign key (tid) references "user"(id);

alter table deed add column tid ksuid;
update deed d set tid = c.tid from company c where d.company_id = c.id;
alter table deed alter tid set not null;
alter table deed add constraint deed_tid_fk_user_id foreign key (tid) references "user"(id);

alter table drain add column tid ksuid;
update drain d set tid = e.tid from entry e where d.entry_id = e.id;
alter table drain alter tid set not null;
alter table drain add constraint drain_tid_fk_user_id foreign key (tid) references "user"(id);

alter table company alter tid set not null;
alter table entry_type alter tid set not null;

alter table entry_type alter tid set default get_tenent();
alter table entry alter tid set default get_tenent();
alter table deed alter tid set default get_tenent();
alter table drain alter tid set default get_tenent();
alter table company alter tid set default get_tenent();

drop policy if exists company_tent on company;
create policy company_tent on company to dots_api_user 
using (tid=get_tenent());

drop policy if exists entry_tent on entry;
create policy entry_tent on entry to dots_api_user 
using (tid=get_tenent());

drop policy if exists entry_type_tent on entry_type;
create policy entry_type_tent on entry_type to dots_api_user 
using (tid=get_tenent());

drop policy if exists deed_tent on deed;
create policy deed_tent on deed to dots_api_user 
using (tid=get_tenent());

drop policy if exists drain_tent on drain;
create policy drain_tent on drain to dots_api_user 
using (tid=get_tenent());
