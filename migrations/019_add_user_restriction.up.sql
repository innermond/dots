alter table package add column field_len jsonb;
update package set field_len = '{
"company": {"longname": 30, "tin": 20, "rn": 20},
"deed": {"title": 15, "unit": 3},
"entry_type": {"code": 10, "description": 30, "unit": 3}
}'::jsonb where name = 'one eye';
update package set field_len = '{
"company": {"longname": 60, "tin": 40, "rn": 40},
"deed": {"title": 30, "unit": 6},
"entry_type": {"code": 20, "description": 60, "unit": 6}
}'::jsonb where name = 'two eyes';
update package set field_len = '{
"company": {"longname": 120, "tin": 80, "rn": 80},
"deed": {"title": 60, "unit": 12},
"entry_type": {"code": 40, "description": 120, "unit": 12}
}'::jsonb where name = 'three eyes';
alter table package alter column field_len set not null;
alter table package add constraint package_field_len_unique unique (field_len);

alter table user_num_record rename to user_restriction;
alter table user_restriction 
  add column field_len jsonb,
  drop constraint user_num_record_user_id_key,
  add constraint user_restriction_user_id_key unique (user_id);
