CREATE TYPE online.payment_provider AS ENUM ('prodamus');

CREATE TABLE online.customer_payment (
  provider online.payment_provider NOT NULL,
  external_id TEXT NOT NULL,
  user_id BIGINT REFERENCES online."user"(id) ON DELETE CASCADE NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  PRIMARY KEY (provider, external_id)
);
