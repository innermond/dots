CREATE TABLE company (
	id int4 primary key generated always as identity,
	tid int4 NOT NULL,
	longname varchar NOT NULL,
	tin varchar NOT NULL,
	rn varchar NOT NULL,
	UNIQUE (tid, rn, tin)
);

CREATE TABLE deed (
	id int8 primary key generated always as identity,
	company_id int4 NULL,
	title varchar NOT NULL,
	quantity float8 NOT NULL DEFAULT 1,
	unit varchar NOT NULL DEFAULT 'pcs'::character varying,
	unitprice numeric(15, 2) NULL
);

CREATE TABLE entry_type (
	id int4 primary key generated always as identity,
	code varchar NOT NULL,
	description text NULL,
	unit varchar NOT NULL,
	tid int4 NOT NULL,
  unique (code, tid)
);

CREATE TABLE entry (
	id int8 primary key generated always as identity,
	entry_type_id int4 NOT NULL,
	date_added timestamptz NULL DEFAULT now(),
	quantity float8 NOT NULL DEFAULT 0.0,
	company_id int4 NOT NULL
);

CREATE TABLE drain (
	deed_id int8 NOT NULL,
	entry_id int8 NOT NULL,
	quantity float8 NOT NULL DEFAULT 0
);
