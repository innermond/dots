alter table company add CONSTRAINT company_tid_fk_user_tid FOREIGN KEY (tid) REFERENCES "user"(id);
alter table deed add CONSTRAINT deed_company_id_fk_company_id FOREIGN KEY (company_id) REFERENCES company(id);
alter table entry_type add CONSTRAINT entry_type_tid_fk_user_id FOREIGN KEY (tid) REFERENCES "user"(id);
alter table entry add CONSTRAINT entry_company_id_fk_company_id FOREIGN KEY (company_id) REFERENCES company(id);
alter table entry add CONSTRAINT entry_entry_type_id_fk_entry_type_id FOREIGN KEY (entry_type_id) REFERENCES entry_type(id);
alter table drain add CONSTRAINT drain_entry_id_fk_entry_id FOREIGN KEY (entry_id) REFERENCES entry(id);
alter table drain add CONSTRAINT drain_deed_id_fk_deed_id FOREIGN KEY (deed_id) REFERENCES deed(id);
