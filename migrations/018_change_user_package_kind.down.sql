alter table "user" drop constraint user_package_kind_fkey;
alter table "user" add constraint user_package_kind_check CHECK ((package_kind = ANY (ARRAY['one eye'::text, 'two eyes'::text, 'three eyes'::text])));

