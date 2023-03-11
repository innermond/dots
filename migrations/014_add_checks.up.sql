alter table entry_type add constraint entry_type_code_check
check (
length(code) > 0
);
