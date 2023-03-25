alter table package drop column field_len;

alter table user_restriction drop column field_len;
alter table user_restriction rename to user_num_record;
alter table user_num_record 
  drop constraint user_restriction_user_id_key,
  add constraint user_num_record_user_id_key unique (user_id);
