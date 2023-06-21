alter table "user" drop constraint user_package_kind_check;
alter table "user" add constraint user_package_kind_fkey foreign key (package_kind) references package (name);
