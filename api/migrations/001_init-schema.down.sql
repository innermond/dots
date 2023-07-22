drop schema if exists api cascade;
drop schema if exists mock cascade;

--
-- Name: api; Type: SCHEMA; Schema: -; Owner: dots_owner
--

CREATE SCHEMA api;

grant usage, create on schema api to dots_readwrite;
grant select, insert, update, delete on all tables in schema api to dots_readwrite;
alter default privileges in schema api grant select, insert, update, delete on tables to dots_readwrite;
grant usage on all sequences in schema api to dots_readwrite;
alter default privileges in schema api grant usage on sequences to dots_readwrite;

--
-- Name: mock; Type: SCHEMA; Schema: -; Owner: dots_owner
--

CREATE SCHEMA mock;
