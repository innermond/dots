alter table api."user" add column pass_hash varchar(255) null;
create unique index user_email_pass_hash on api."user" (email, pass_hash);
