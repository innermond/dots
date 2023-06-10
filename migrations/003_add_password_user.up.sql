alter table "user" add column pass_hash varchar(255) null;
create unique index user_email_pass_hash on "user" (email, pass_hash);
