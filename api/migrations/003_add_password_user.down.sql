alter table api."user" drop column pass_hash;
drop index if exists user_email_pass_hash;
