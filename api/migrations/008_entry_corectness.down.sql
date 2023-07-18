drop trigger if exists entry_type_company_has_same_tid_tg on entry;
drop function if exists entry_type_company_same_tid;
alter table entry drop constraint check_entry_quantity;
alter table entry alter column quantity set default 0.0;
alter table entry alter column tid drop not null;
