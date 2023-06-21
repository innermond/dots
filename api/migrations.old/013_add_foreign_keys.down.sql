alter table company drop CONSTRAINT company_tid_fk_user_tid;
alter table deed drop CONSTRAINT deed_company_id_fk_company_id;
alter table entry_type drop CONSTRAINT entry_type_tid_fk_user_id;
alter table entry drop CONSTRAINT entry_company_id_fk_company_id;
alter table entry drop CONSTRAINT entry_entry_type_id_fk_entry_type_id;
alter table drain drop CONSTRAINT drain_entry_id_fk_entry_id;
alter table drain drop CONSTRAINT drain_deed_id_fk_deed_id;

