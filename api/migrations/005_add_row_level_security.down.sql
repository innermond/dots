drop policy if exists company_tent on company;
drop policy if exists entry_tent on entry;
drop policy if exists entry_type_tent on entry_type;
drop policy if exists deed_tent on deed;
drop policy if exists drain_tent on drain;

alter table entry drop constraint entry_tid_fk_user_id;
alter table deed drop constraint deed_tid_fk_user_id;
alter table drain drop constraint drain_tid_fk_user_id;

alter table entry drop column tid;
alter table deed drop column tid;
alter table drain drop column tid;

alter table company disable row level security;
alter table entry_type disable row level security;
alter table entry disable row level security;
alter table deed disable row level security;
alter table drain disable row level security;
