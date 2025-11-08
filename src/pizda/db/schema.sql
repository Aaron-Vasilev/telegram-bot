CREATE SCHEMA pizda;

ALTER SCHEMA pizda OWNER TO postgres;

CREATE TABLE pizda."user" (
  id BIGINT PRIMARY KEY,
  username CHARACTER VARYING(32) NOT NULL DEFAULT '',
  first_name character varying(128) NOT NULL,
  last_name character varying(128) NOT NULL DEFAULT '',
  is_blocked boolean DEFAULT false
);

CREATE TYPE pizda.payment_method AS ENUM ('MIR', 'BIT', 'PAYPAL');

CREATE TABLE pizda.payment (
  id SERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES pizda."user"(id) ON DELETE CASCADE,
  method pizda.payment_method NOT NULL,
  creation_date DATE NOT NULL DEFAULT NOW(),
  period daterange DEFAULT daterange(CURRENT_DATE, (CURRENT_DATE + INTERVAL '2 months')::date, '[]')
);

CREATE TABLE pizda.file (
  id SERIAL PRIMARY KEY,
  file_id TEXT NOT NULL,
  name TEXT NOT NULL UNIQUE
);
