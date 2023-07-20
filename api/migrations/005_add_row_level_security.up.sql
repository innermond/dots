create or replace function api.get_tenent() returns api.ksuid as $$
begin
	return nullif(current_setting('app.uid', true), '');
end;
$$ language plpgsql;

alter table api.entry add column tid api.ksuid;
update api.entry e set tid = c.tid from api.company c where c.id = e.company_id;
alter table api.entry alter tid set default not null;
alter table api.entry add constraint entry_tid_fk_user_id foreign key (tid) references api."user"(id);

alter table api.deed add column tid api.ksuid;
update api.deed d set tid = c.tid from api.company c where d.company_id = c.id;
alter table api.deed alter tid set not null;
alter table api.deed add constraint deed_tid_fk_user_id foreign key (tid) references api."user"(id);

alter table api.drain add column tid api.ksuid;
update api.drain d set tid = e.tid from api.entry e where d.entry_id = e.id;
alter table api.drain alter tid set not null;
alter table api.drain add constraint drain_tid_fk_user_id foreign key (tid) references api."user"(id);

alter table api.company alter tid set not null;
alter table api.entry_type alter tid set not null;

alter table api.company alter tid set default api.get_tenent();
alter table api.company enable row level security;
alter table api.entry_type alter tid set default api.get_tenent();
alter table api.entry_type enable row level security;
alter table api.entry alter tid set default api.get_tenent();
alter table api.entry enable row level security;
alter table api.deed alter tid set default api.get_tenent();
alter table api.deed enable row level security;
alter table api.drain alter tid set default api.get_tenent();
alter table api.drain enable row level security;

drop policy if exists company_tent on api.company;
create policy company_tent on api.company to dots_api_user 
using (tid=api.get_tenent());

drop policy if exists entry_tent on api.entry;
create policy entry_tent on api.entry to dots_api_user 
using (tid=api.get_tenent());

drop policy if exists entry_type_tent on api.entry_type;
create policy entry_type_tent on api.entry_type to dots_api_user 
using (tid=api.get_tenent());

drop policy if exists deed_tent on api.deed;
create policy deed_tent on api.deed to dots_api_user 
using (tid=api.get_tenent());

drop policy if exists drain_tent on api.drain;
create policy drain_tent on api.drain to dots_api_user 
using (tid=api.get_tenent());
