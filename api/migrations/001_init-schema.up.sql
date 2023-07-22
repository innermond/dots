--
-- PostgreSQL database dump
--

-- Dumped from database version 15.1 (Debian 15.1-1.pgdg110+1)
-- Dumped by pg_dump version 15.1 (Debian 15.1-1.pgdg110+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: ksuid; Type: DOMAIN; Schema: api; Owner: dots_owner
--

CREATE DOMAIN api.ksuid AS character varying(27);


ALTER DOMAIN api.ksuid OWNER TO dots_owner;

--
-- Name: company_has_same_tid(); Type: FUNCTION; Schema: api; Owner: dots_owner
--

CREATE FUNCTION api.company_has_same_tid() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
declare has_same boolean;
begin
	select exists(select id from api.company c where c.id=NEW.company_id and c.tid=NEW.tid) into has_same;
	if not has_same then
		raise exception 'company % has not the same tid % or not exists', NEW.company_id, NEW.tid;
	end if;
	return NEW;
end;
$$;


ALTER FUNCTION api.company_has_same_tid() OWNER TO dots_owner;

--
-- Name: deed_entry_same_tid(); Type: FUNCTION; Schema: api; Owner: dots_owner
--

CREATE FUNCTION api.deed_entry_same_tid() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
declare 
	has_not_same boolean;
	deed_tid api.ksuid;
	entry_tid api.ksuid;
begin
	select tid from api.deed where id = NEW.deed_id into deed_tid;
	select tid from api.entry where id = NEW.entry_id into entry_tid;
	has_not_same = (
		(deed_tid is not null and entry_tid is not null) and 
		(deed_tid != entry_tid or (deed_tid != NEW.tid or entry_tid != NEW.tid))
	);
	if  has_not_same then 
		raise exception 'deed % and entry % has not the same expected tid %', deed_tid, entry_tid, NEW.tid; 	
	end if;
	return NEW;
end;
$$;


ALTER FUNCTION api.deed_entry_same_tid() OWNER TO dots_owner;

--
-- Name: entry_type_company_same_tid(); Type: FUNCTION; Schema: api; Owner: dots_owner
--

CREATE FUNCTION api.entry_type_company_same_tid() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
declare
	company_tid api.ksuid;
	entry_type_tid api.ksuid;
	uid api.ksuid;
begin
	select id from api."user" where id = NEW.tid into uid;
	if uid is null then
		raise exception 'not found uid';
	end if;

	if NEW.company_id is not null then
		select tid from api.company where id = NEW.company_id into company_tid;
		if company_tid != NEW.tid then
			raise exception 'company % has not expected tid %', company_tid, NEW.tid;
		end if;
	end if;

	if NEW.entry_type_id is not null then
		select tid from api.entry_type where id = NEW.entry_type_id into entry_type_tid;
		if entry_type_tid != NEW.tid then
			raise exception 'entry type % has not expected tid %', entry_type_tid, NEW.tid;
		end if;
	end if;

	if  company_tid != entry_type_tid then
		raise exception 'company % and entry type % has not the same expected tid %', company_tid, entry_type_tid, NEW.tid;
	end if;

	raise notice 'new tid: % old tid %', NEW.tid, OLD.tid;
	return NEW;
end;
$$;


ALTER FUNCTION api.entry_type_company_same_tid() OWNER TO dots_owner;

--
-- Name: get_tenent(); Type: FUNCTION; Schema: api; Owner: dots_owner
--

CREATE FUNCTION api.get_tenent() RETURNS api.ksuid
    LANGUAGE plpgsql
    AS $$
begin
	return nullif(current_setting('app.uid', true), '');
end;
$$;


ALTER FUNCTION api.get_tenent() OWNER TO dots_owner;

--
-- Name: ksuid(); Type: FUNCTION; Schema: api; Owner: dots_owner
--

CREATE FUNCTION api.ksuid() RETURNS text
    LANGUAGE plpgsql
    AS $$
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

end $$;


ALTER FUNCTION api.ksuid() OWNER TO dots_owner;

--
-- Name: update_drain_is_deleted(); Type: FUNCTION; Schema: api; Owner: dots_owner
--

CREATE FUNCTION api.update_drain_is_deleted() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
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
$$;


ALTER FUNCTION api.update_drain_is_deleted() OWNER TO dots_owner;

--
-- Name: create_company(); Type: FUNCTION; Schema: mock; Owner: dots_owner
--

CREATE FUNCTION mock.create_company() RETURNS void
    LANGUAGE sql
    AS $$
	insert into api.company
	(longname, tin, rn, tid)
	values 
	(
		mock.generate_random_title(3, 4), 
		substr(md5(random()::text), 1, 12), 
		substr(md5(random()::text), 1, 8), 
		mock.gen_rnd_tid()
	)
$$;


ALTER FUNCTION mock.create_company() OWNER TO dots_owner;

--
-- Name: create_entry(boolean, integer, integer, integer, integer); Type: FUNCTION; Schema: mock; Owner: dots_owner
--

CREATE FUNCTION mock.create_entry(dispersed boolean DEFAULT true, entry_type_id integer DEFAULT NULL::integer, company_id integer DEFAULT NULL::integer, a integer DEFAULT 1, z integer DEFAULT 1000) RETURNS void
    LANGUAGE plpgsql
    AS $$
declare vtid api.ksuid;
declare etid integer default null;
declare cid integer default null;
begin
	select mock.gen_rnd_tid() into vtid;
	if dispersed then
		insert into api.entry
		(entry_type_id, quantity, company_id, tid)
		values 
		(
			(select id from api.entry_type et where et.tid = vtid order by random() limit 1),
			(select floor(random()*(z-a+1))+a),
			(select id from api.company c where c.tid = vtid order by random() limit 1),
			vtid
		);
	else
		select id, et.tid into etid, vtid from api.entry_type et where id = entry_type_id;
		select id into cid from api.company c where id = company_id and c.tid = vtid;
		if etid is null or cid is null then
			raise exception 'entry type %', entry_type_id;
		end if;
		insert into api.entry
		(entry_type_id, quantity, company_id, tid)
		values 
		(
			etid,
			(select floor(random()*(z-a+1))+a),
			cid,
			vtid
		);		
	
	end if;
end;
$$;


ALTER FUNCTION mock.create_entry(dispersed boolean, entry_type_id integer, company_id integer, a integer, z integer) OWNER TO dots_owner;

--
-- Name: create_entry_type(); Type: FUNCTION; Schema: mock; Owner: dots_owner
--

CREATE FUNCTION mock.create_entry_type() RETURNS void
    LANGUAGE sql
    AS $$
	insert into api.entry_type
	(code, description, unit, tid)
	values 
	(
		concat(substr(md5(random()::text), 1, 6),'-', substr(md5(random()::text), 1, 4)),
		mock.generate_random_title(3, 10),
		mock.gen_rnd_unit(),
		mock.gen_rnd_tid()
	)
$$;


ALTER FUNCTION mock.create_entry_type() OWNER TO dots_owner;

--
-- Name: gen_rnd_tid(); Type: FUNCTION; Schema: mock; Owner: dots_owner
--

CREATE FUNCTION mock.gen_rnd_tid() RETURNS text
    LANGUAGE sql
    AS $$
	select id from (values ('2PH25DxmohuFCf3w73fQSTLJeVO'), ('2PH24UhBlN5tlYdAmpdwiyPuWgB')) a(id) order by random() limit 1;
$$;


ALTER FUNCTION mock.gen_rnd_tid() OWNER TO dots_owner;

--
-- Name: gen_rnd_unit(); Type: FUNCTION; Schema: mock; Owner: dots_owner
--

CREATE FUNCTION mock.gen_rnd_unit() RETURNS text
    LANGUAGE sql
    AS $$
	select id from (values ('unit'), ('buc'), ('piece'), ('cm'), ('mm'), ('pcs'), ('sq'), ('square meter'), ('piese'), ('hour'), ('rola')) a(id) order by random() limit 1;
$$;


ALTER FUNCTION mock.gen_rnd_unit() OWNER TO dots_owner;

--
-- Name: generate_random_title(integer, integer); Type: FUNCTION; Schema: mock; Owner: dots_owner
--

CREATE FUNCTION mock.generate_random_title(a integer DEFAULT 1, z integer DEFAULT 4) RETURNS text
    LANGUAGE sql
    AS $$
  select initcap(array_to_string(array(
    select v from mock.word order by random() limit (select floor(random()*(z-a+1)+a)::int)
  ), ' '))
$$;


ALTER FUNCTION mock.generate_random_title(a integer, z integer) OWNER TO dots_owner;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: auth; Type: TABLE; Schema: api; Owner: dots_owner
--

CREATE TABLE api.auth (
    id integer NOT NULL,
    user_id api.ksuid NOT NULL,
    source text NOT NULL,
    source_id text NOT NULL,
    access_token text NOT NULL,
    refresh_token text NOT NULL,
    expiry timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


ALTER TABLE api.auth OWNER TO dots_owner;

--
-- Name: auth_id_seq; Type: SEQUENCE; Schema: api; Owner: dots_owner
--

ALTER TABLE api.auth ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME api.auth_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: company; Type: TABLE; Schema: api; Owner: dots_owner
--

CREATE TABLE api.company (
    id integer NOT NULL,
    tid api.ksuid DEFAULT api.get_tenent() NOT NULL,
    longname character varying NOT NULL,
    tin character varying NOT NULL,
    rn character varying NOT NULL,
    deleted_at timestamp with time zone
);


ALTER TABLE api.company OWNER TO dots_owner;

--
-- Name: company_id_seq; Type: SEQUENCE; Schema: api; Owner: dots_owner
--

ALTER TABLE api.company ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME api.company_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: deed; Type: TABLE; Schema: api; Owner: dots_owner
--

CREATE TABLE api.deed (
    id bigint NOT NULL,
    company_id integer,
    title character varying NOT NULL,
    quantity double precision DEFAULT 1 NOT NULL,
    unit character varying DEFAULT 'pcs'::character varying NOT NULL,
    unitprice numeric(15,2),
    deleted_at timestamp with time zone,
    tid api.ksuid DEFAULT api.get_tenent() NOT NULL
);


ALTER TABLE api.deed OWNER TO dots_owner;

--
-- Name: deed_id_seq; Type: SEQUENCE; Schema: api; Owner: dots_owner
--

ALTER TABLE api.deed ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME api.deed_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: drain; Type: TABLE; Schema: api; Owner: dots_owner
--

CREATE TABLE api.drain (
    deed_id bigint NOT NULL,
    entry_id bigint NOT NULL,
    quantity double precision DEFAULT 0 NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    tid api.ksuid DEFAULT api.get_tenent() NOT NULL
);


ALTER TABLE api.drain OWNER TO dots_owner;

--
-- Name: entry; Type: TABLE; Schema: api; Owner: dots_owner
--

CREATE TABLE api.entry (
    id bigint NOT NULL,
    entry_type_id integer NOT NULL,
    date_added timestamp with time zone DEFAULT now(),
    quantity double precision NOT NULL,
    company_id integer NOT NULL,
    deleted_at timestamp with time zone,
    tid api.ksuid DEFAULT api.get_tenent() NOT NULL,
    CONSTRAINT check_entry_quantity CHECK ((quantity > (0)::double precision))
);


ALTER TABLE api.entry OWNER TO dots_owner;

--
-- Name: entry_id_seq; Type: SEQUENCE; Schema: api; Owner: dots_owner
--

ALTER TABLE api.entry ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME api.entry_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: entry_type; Type: TABLE; Schema: api; Owner: dots_owner
--

CREATE TABLE api.entry_type (
    id integer NOT NULL,
    code character varying NOT NULL,
    description text,
    unit character varying NOT NULL,
    tid api.ksuid DEFAULT api.get_tenent() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT entry_type_code_check CHECK ((length((code)::text) > 0))
);


ALTER TABLE api.entry_type OWNER TO dots_owner;

--
-- Name: entry_type_id_seq; Type: SEQUENCE; Schema: api; Owner: dots_owner
--

ALTER TABLE api.entry_type ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME api.entry_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: package; Type: TABLE; Schema: api; Owner: dots_owner
--

CREATE TABLE api.package (
    name text NOT NULL,
    company smallint NOT NULL,
    deed integer NOT NULL,
    drain bigint NOT NULL,
    entry_type integer NOT NULL,
    entry bigint NOT NULL,
    field_len jsonb NOT NULL,
    CONSTRAINT package_name_check CHECK ((name = ANY (ARRAY['one eye'::text, 'two eyes'::text, 'three eyes'::text])))
);


ALTER TABLE api.package OWNER TO dots_owner;

--
-- Name: user; Type: TABLE; Schema: api; Owner: dots_owner
--

CREATE TABLE api."user" (
    id api.ksuid DEFAULT api.ksuid() NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    email text,
    api_key text NOT NULL,
    package_kind text DEFAULT 'one eye'::text NOT NULL,
    powers text[],
    pass_hash character varying(255)
);


ALTER TABLE api."user" OWNER TO dots_owner;

--
-- Name: user_restriction; Type: TABLE; Schema: api; Owner: dots_owner
--

CREATE TABLE api.user_restriction (
    user_id api.ksuid NOT NULL,
    company smallint,
    deed integer,
    drain bigint,
    entry_type integer,
    entry bigint,
    field_len jsonb
);


ALTER TABLE api.user_restriction OWNER TO dots_owner;

--
-- Name: word; Type: TABLE; Schema: mock; Owner: dots_owner
--

CREATE TABLE mock.word (
    v text
);


ALTER TABLE mock.word OWNER TO dots_owner;

--
-- Name: auth auth_pkey; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.auth
    ADD CONSTRAINT auth_pkey PRIMARY KEY (id);


--
-- Name: auth auth_source_source_id; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.auth
    ADD CONSTRAINT auth_source_source_id UNIQUE (source, source_id);


--
-- Name: auth auth_user_id_source; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.auth
    ADD CONSTRAINT auth_user_id_source UNIQUE (user_id, source);


--
-- Name: company company_pkey; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.company
    ADD CONSTRAINT company_pkey PRIMARY KEY (id);


--
-- Name: company company_tid_rn_tin_key; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.company
    ADD CONSTRAINT company_tid_rn_tin_key UNIQUE (tid, rn, tin);


--
-- Name: deed deed_pkey; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.deed
    ADD CONSTRAINT deed_pkey PRIMARY KEY (id);


--
-- Name: drain drain_deed_entry_unique_key; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.drain
    ADD CONSTRAINT drain_deed_entry_unique_key UNIQUE (deed_id, entry_id);


--
-- Name: entry entry_pkey; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.entry
    ADD CONSTRAINT entry_pkey PRIMARY KEY (id);


--
-- Name: entry_type entry_type_code_tid_key; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.entry_type
    ADD CONSTRAINT entry_type_code_tid_key UNIQUE (code, tid);


--
-- Name: entry_type entry_type_pkey; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.entry_type
    ADD CONSTRAINT entry_type_pkey PRIMARY KEY (id);


--
-- Name: package package_field_len_unique; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.package
    ADD CONSTRAINT package_field_len_unique UNIQUE (field_len);


--
-- Name: package package_name_key; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.package
    ADD CONSTRAINT package_name_key UNIQUE (name);


--
-- Name: user user_api_key_key; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api."user"
    ADD CONSTRAINT user_api_key_key UNIQUE (api_key);


--
-- Name: user user_email_key; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api."user"
    ADD CONSTRAINT user_email_key UNIQUE (email);


--
-- Name: user_restriction user_restriction_user_id_key; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.user_restriction
    ADD CONSTRAINT user_restriction_user_id_key UNIQUE (user_id);


--
-- Name: user users_pkey; Type: CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api."user"
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

--
-- Name: user_email_pass_hash; Type: INDEX; Schema: api; Owner: dots_owner
--

CREATE UNIQUE INDEX user_email_pass_hash ON api."user" USING btree (email, pass_hash);


--
-- Name: deed company_has_same_tid_tg; Type: TRIGGER; Schema: api; Owner: dots_owner
--

CREATE TRIGGER company_has_same_tid_tg BEFORE INSERT OR UPDATE ON api.deed FOR EACH ROW EXECUTE FUNCTION api.company_has_same_tid();


--
-- Name: drain deed_entry_has_same_tid_tg; Type: TRIGGER; Schema: api; Owner: dots_owner
--

CREATE TRIGGER deed_entry_has_same_tid_tg BEFORE INSERT OR UPDATE ON api.drain FOR EACH ROW EXECUTE FUNCTION api.deed_entry_same_tid();


--
-- Name: entry entry_type_company_has_same_tid_tg; Type: TRIGGER; Schema: api; Owner: dots_owner
--

CREATE TRIGGER entry_type_company_has_same_tid_tg BEFORE INSERT OR UPDATE ON api.entry FOR EACH ROW EXECUTE FUNCTION api.entry_type_company_same_tid();


--
-- Name: auth auth_user_id_fkey; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.auth
    ADD CONSTRAINT auth_user_id_fkey FOREIGN KEY (user_id) REFERENCES api."user"(id);


--
-- Name: company company_tid_fk_user_tid; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.company
    ADD CONSTRAINT company_tid_fk_user_tid FOREIGN KEY (tid) REFERENCES api."user"(id);


--
-- Name: deed deed_company_id_fk_company_id; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.deed
    ADD CONSTRAINT deed_company_id_fk_company_id FOREIGN KEY (company_id) REFERENCES api.company(id);


--
-- Name: deed deed_tid_fk_user_id; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.deed
    ADD CONSTRAINT deed_tid_fk_user_id FOREIGN KEY (tid) REFERENCES api."user"(id);


--
-- Name: drain drain_deed_id_fk_deed_id; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.drain
    ADD CONSTRAINT drain_deed_id_fk_deed_id FOREIGN KEY (deed_id) REFERENCES api.deed(id) ON UPDATE CASCADE;


--
-- Name: drain drain_entry_id_fk_entry_id; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.drain
    ADD CONSTRAINT drain_entry_id_fk_entry_id FOREIGN KEY (entry_id) REFERENCES api.entry(id);


--
-- Name: drain drain_tid_fk_user_id; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.drain
    ADD CONSTRAINT drain_tid_fk_user_id FOREIGN KEY (tid) REFERENCES api."user"(id);


--
-- Name: entry entry_company_id_fk_company_id; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.entry
    ADD CONSTRAINT entry_company_id_fk_company_id FOREIGN KEY (company_id) REFERENCES api.company(id);


--
-- Name: entry entry_entry_type_id_fk_entry_type_id; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.entry
    ADD CONSTRAINT entry_entry_type_id_fk_entry_type_id FOREIGN KEY (entry_type_id) REFERENCES api.entry_type(id);


--
-- Name: entry entry_tid_fk_user_id; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.entry
    ADD CONSTRAINT entry_tid_fk_user_id FOREIGN KEY (tid) REFERENCES api."user"(id);


--
-- Name: entry_type entry_type_tid_fk_user_id; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.entry_type
    ADD CONSTRAINT entry_type_tid_fk_user_id FOREIGN KEY (tid) REFERENCES api."user"(id);


--
-- Name: user_restriction user_num_record_user_id_fkey; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api.user_restriction
    ADD CONSTRAINT user_num_record_user_id_fkey FOREIGN KEY (user_id) REFERENCES api."user"(id) DEFERRABLE;


--
-- Name: user user_package_kind_fkey; Type: FK CONSTRAINT; Schema: api; Owner: dots_owner
--

ALTER TABLE ONLY api."user"
    ADD CONSTRAINT user_package_kind_fkey FOREIGN KEY (package_kind) REFERENCES api.package(name);


--
-- Name: company; Type: ROW SECURITY; Schema: api; Owner: dots_owner
--

ALTER TABLE api.company ENABLE ROW LEVEL SECURITY;

--
-- Name: company company_tent; Type: POLICY; Schema: api; Owner: dots_owner
--

CREATE POLICY company_tent ON api.company TO dots_api_user USING (((tid)::text = (api.get_tenent())::text));


--
-- Name: deed; Type: ROW SECURITY; Schema: api; Owner: dots_owner
--

ALTER TABLE api.deed ENABLE ROW LEVEL SECURITY;

--
-- Name: deed deed_tent; Type: POLICY; Schema: api; Owner: dots_owner
--

CREATE POLICY deed_tent ON api.deed TO dots_api_user USING (((tid)::text = (api.get_tenent())::text));


--
-- Name: drain; Type: ROW SECURITY; Schema: api; Owner: dots_owner
--

ALTER TABLE api.drain ENABLE ROW LEVEL SECURITY;

--
-- Name: drain drain_tent; Type: POLICY; Schema: api; Owner: dots_owner
--

CREATE POLICY drain_tent ON api.drain TO dots_api_user USING (((tid)::text = (api.get_tenent())::text));


--
-- Name: entry; Type: ROW SECURITY; Schema: api; Owner: dots_owner
--

ALTER TABLE api.entry ENABLE ROW LEVEL SECURITY;

--
-- Name: entry entry_tent; Type: POLICY; Schema: api; Owner: dots_owner
--

CREATE POLICY entry_tent ON api.entry TO dots_api_user USING (((tid)::text = (api.get_tenent())::text));


--
-- Name: entry_type; Type: ROW SECURITY; Schema: api; Owner: dots_owner
--

ALTER TABLE api.entry_type ENABLE ROW LEVEL SECURITY;

--
-- Name: entry_type entry_type_tent; Type: POLICY; Schema: api; Owner: dots_owner
--

CREATE POLICY entry_type_tent ON api.entry_type TO dots_api_user USING (((tid)::text = (api.get_tenent())::text));
