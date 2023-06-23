DROP TABLE IF EXISTS public.auth CASCADE;

DROP TABLE IF EXISTS public.company CASCADE;

DROP TABLE IF EXISTS public.deed CASCADE;

DROP TABLE IF EXISTS public.drain CASCADE;

DROP TABLE IF EXISTS public.entry CASCADE;

DROP TABLE IF EXISTS public.entry_type CASCADE;

DROP TABLE IF EXISTS public.package CASCADE;

DROP TABLE IF EXISTS public."user" CASCADE;

DROP TABLE IF EXISTS public.user_restriction CASCADE;

DROP DOMAIN IF EXISTS public.KSUID;

CREATE DOMAIN KSUID CHARACTER VARYING(27);

CREATE OR REPLACE FUNCTION public.ksuid()
 RETURNS text
 LANGUAGE plpgsql
AS $function$
declare
	v_time timestamp with time zone := null;
	v_seconds numeric(50) := null;
	v_numeric numeric(50) := null;
	v_epoch numeric(50) = 1400000000; -- 2014-05-13T16:53:20Z
	v_base62 text := '';
	v_alphabet char array[62] := array[
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
		'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
		'U', 'V', 'W', 'X', 'Y', 'Z',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
		'u', 'v', 'w', 'x', 'y', 'z'];
	i integer := 0;
begin

	-- Get the current time
	v_time := clock_timestamp();

	-- Extract epoch seconds
	v_seconds := EXTRACT(EPOCH FROM v_time) - v_epoch;

	-- Generate a KSUID in a numeric variable
	v_numeric := v_seconds * pow(2::numeric(50), 128) -- 32 bits for seconds and 128 bits for randomness
		+ ((random()::numeric(70,20) * pow(2::numeric(70,20), 48))::numeric(50) * pow(2::numeric(50), 80)::numeric(50))
		+ ((random()::numeric(70,20) * pow(2::numeric(70,20), 40))::numeric(50) * pow(2::numeric(50), 40)::numeric(50))
		+  (random()::numeric(70,20) * pow(2::numeric(70,20), 40))::numeric(50);

	-- Encode it to base-62
	while v_numeric <> 0 loop
		v_base62 := v_base62 || v_alphabet[mod(v_numeric, 62) + 1];
		v_numeric := div(v_numeric, 62);
	end loop;
	v_base62 := reverse(v_base62);
	v_base62 := lpad(v_base62, 27, '0');

	return v_base62;

end $function$
;

CREATE OR REPLACE FUNCTION public.update_drain_is_deleted()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
declare var_is_deleted boolean;
begin
  if (NEW.deleted_at is null) then
   var_is_deleted := false;
  else
    var_is_deleted := true;
  end if;
  update drain set is_deleted = var_is_deleted where deed_id = NEW.id;
  return NEW;
end;
$function$
;


-- public.package definition

-- Drop table

-- DROP TABLE IF EXISTS package;

CREATE TABLE IF NOT EXISTS package (
	"name" text NOT NULL,
	company int2 NOT NULL,
	deed int4 NOT NULL,
	drain int8 NOT NULL,
	entry_type int4 NOT NULL,
	entry int8 NOT NULL,
	field_len jsonb NOT NULL,
	CONSTRAINT package_field_len_unique UNIQUE (field_len),
	CONSTRAINT package_name_check CHECK ((name = ANY (ARRAY['one eye'::text, 'two eyes'::text, 'three eyes'::text]))),
	CONSTRAINT package_name_key UNIQUE (name)
);


-- public."user" definition

-- Drop table

-- DROP TABLE IF EXISTS "user";

CREATE TABLE IF NOT EXISTS "user" (
	id public.KSUID NOT NULL DEFAULT ksuid(),
	"name" text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NULL,
	email text NULL,
	api_key text NOT NULL,
	package_kind text NOT NULL DEFAULT 'one eye'::text,
	powers _text NULL,
	CONSTRAINT user_api_key_key UNIQUE (api_key),
	CONSTRAINT user_email_key UNIQUE (email),
	CONSTRAINT users_pkey PRIMARY KEY (id),
	CONSTRAINT user_package_kind_fkey FOREIGN KEY (package_kind) REFERENCES package("name")
);


-- public.user_restriction definition

-- Drop table

-- DROP TABLE IF EXISTS user_restriction;

CREATE TABLE IF NOT EXISTS user_restriction (
	user_id public.KSUID NOT NULL,
	company int2 NULL,
	deed int4 NULL,
	drain int8 NULL,
	entry_type int4 NULL,
	entry int8 NULL,
	field_len jsonb NULL,
	CONSTRAINT user_restriction_user_id_key UNIQUE (user_id),
	CONSTRAINT user_num_record_user_id_fkey FOREIGN KEY (user_id) REFERENCES "user"(id) DEFERRABLE
);


-- public.auth definition

-- Drop table

-- DROP TABLE IF EXISTS auth;

CREATE TABLE IF NOT EXISTS auth (
	id int4 NOT NULL GENERATED ALWAYS AS IDENTITY,
	user_id public.KSUID NOT NULL,
	"source" text NOT NULL,
	source_id text NOT NULL,
	access_token text NOT NULL,
	refresh_token text NOT NULL,
	expiry timestamptz NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL,
	CONSTRAINT auth_pkey PRIMARY KEY (id),
	CONSTRAINT auth_source_source_id UNIQUE (source, source_id),
	CONSTRAINT auth_user_id_source UNIQUE (user_id, source),
	CONSTRAINT auth_user_id_fkey FOREIGN KEY (user_id) REFERENCES "user"(id)
);


-- public.company definition

-- Drop table

-- DROP TABLE IF EXISTS company;

CREATE TABLE IF NOT EXISTS company (
	id int4 NOT NULL GENERATED ALWAYS AS IDENTITY,
	tid public.KSUID NOT NULL,
	longname varchar NOT NULL,
	tin varchar NOT NULL,
	rn varchar NOT NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT company_pkey PRIMARY KEY (id),
	CONSTRAINT company_tid_rn_tin_key UNIQUE (tid, rn, tin),
	CONSTRAINT company_tid_fk_user_tid FOREIGN KEY (tid) REFERENCES "user"(id)
);


-- public.deed definition

-- Drop table

-- DROP TABLE IF EXISTS deed;

CREATE TABLE IF NOT EXISTS deed (
	id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
	company_id int4 NULL,
	title varchar NOT NULL,
	quantity float8 NOT NULL DEFAULT 1,
	unit varchar NOT NULL DEFAULT 'pcs'::character varying,
	unitprice numeric(15, 2) NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT deed_pkey PRIMARY KEY (id),
	CONSTRAINT deed_company_id_fk_company_id FOREIGN KEY (company_id) REFERENCES company(id)
);

-- Table Triggers

create trigger update_drain_is_deleted_trigger after update on
public.deed for each row
when ((old.deleted_at is distinct
from new.deleted_at)) execute function update_drain_is_deleted();


-- public.entry_type definition

-- Drop table

-- DROP TABLE IF EXISTS entry_type;

CREATE TABLE IF NOT EXISTS entry_type (
	id int4 NOT NULL GENERATED ALWAYS AS IDENTITY,
	code varchar NOT NULL,
	description text NULL,
	unit varchar NOT NULL,
	tid public.KSUID NOT NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT entry_type_code_check CHECK ((length((code)::text) > 0)),
	CONSTRAINT entry_type_code_tid_key UNIQUE (code, tid),
	CONSTRAINT entry_type_pkey PRIMARY KEY (id),
	CONSTRAINT entry_type_tid_fk_user_id FOREIGN KEY (tid) REFERENCES "user"(id)
);


-- public.entry definition

-- Drop table

-- DROP TABLE IF EXISTS entry;

CREATE TABLE IF NOT EXISTS entry (
	id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
	entry_type_id int4 NOT NULL,
	date_added timestamptz NULL DEFAULT now(),
	quantity float8 NOT NULL DEFAULT 0.0,
	company_id int4 NOT NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT entry_pkey PRIMARY KEY (id),
	CONSTRAINT entry_company_id_fk_company_id FOREIGN KEY (company_id) REFERENCES company(id),
	CONSTRAINT entry_entry_type_id_fk_entry_type_id FOREIGN KEY (entry_type_id) REFERENCES entry_type(id)
);


-- public.drain definition

-- Drop table

-- DROP TABLE IF EXISTS drain;

CREATE TABLE IF NOT EXISTS drain (
	deed_id int8 NOT NULL,
	entry_id int8 NOT NULL,
	quantity float8 NOT NULL DEFAULT 0,
	is_deleted bool NOT NULL DEFAULT false,
	CONSTRAINT drain_deed_entry_unique_key UNIQUE (deed_id, entry_id),
	CONSTRAINT drain_deed_id_fk_deed_id FOREIGN KEY (deed_id) REFERENCES deed(id) ON UPDATE CASCADE,
	CONSTRAINT drain_entry_id_fk_entry_id FOREIGN KEY (entry_id) REFERENCES entry(id)
);
