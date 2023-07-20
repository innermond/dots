drop trigger if exists entry_type_company_has_same_tid_tg on api.entry;
drop function if exists api.entry_type_company_same_tid;
alter table api.entry drop constraint check_entry_quantity;
alter table api.entry alter column quantity set default 0.0;
alter table api.entry alter column tid drop not null;
