alter table "auth" alter created_at type timestamp;
alter table "auth" alter created_at drop default;

alter table "auth" alter updated_at type timestamp;
alter table "auth" alter expiry type timestamp;

alter table "user" alter created_at type timestamp;
alter table "user" alter created_at drop default;
alter table "user" alter updated_at type timestamp;
