drop policy if exists company_tent on api.company;
drop policy if exists entry_tent on api.entry;
drop policy if exists entry_type_tent on api.entry_type;
drop policy if exists deed_tent on api.deed;
drop policy if exists drain_tent on api.drain;

alter table api.entry drop constraint entry_tid_fk_user_id;
alter table api.deed drop constraint deed_tid_fk_user_id;
alter table api.drain drop constraint drain_tid_fk_user_id;

alter table api.entry drop column tid;
alter table api.deed drop column tid;
alter table api.drain drop column tid;

alter table api.company disable row level security;
alter table api.entry_type disable row level security;
alter table api.entry disable row level security;
alter table api.deed disable row level security;
alter table api.drain disable row level security;
