alter table "auth" add constraint auth_user_id_source unique (user_id, "source");
alter table "auth" add constraint auth_source_source_id unique ("source", source_id);