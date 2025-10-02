CREATE SCHEMA pizda;

ALTER SCHEMA pizda OWNER TO postgres;

CREATE TABLE pizda."user" (
  id SERIAL PRIMARY KEY,
  tg_id BIGINT NOT NULL UNIQUE,
  username CHARACTER VARYING(32) NOT NULL DEFAULT '',
  first_name character varying(128) NOT NULL,
  last_name character varying(128) NOT NULL DEFAULT '',
  is_blocked boolean DEFAULT false
);

CREATE TYPE pizda.payment_method AS ENUM ('MIR', 'BIT', 'PAYPAL');

CREATE TABLE pizda.payment (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES pizda."user"(id) ON DELETE CASCADE,
  method pizda.payment_method NOT NULL,
  creation_date DATE NOT NULL DEFAULT NOW()
);

ALTER TABLE pizda.payment
ADD COLUMN period daterange DEFAULT daterange(CURRENT_DATE, (CURRENT_DATE + INTERVAL '2 months')::date, '[]');

