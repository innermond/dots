alter table "auth" alter created_at type timestamptz;
alter table "auth" alter created_at set default now();

alter table "auth" alter updated_at type timestamptz;
alter table "auth" alter expiry type timestamptz;

alter table "user" alter created_at type timestamptz;
alter table "user" alter created_at set default now();
alter table "user" alter updated_at type timestamptz;
