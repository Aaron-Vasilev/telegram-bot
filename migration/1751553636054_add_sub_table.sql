CREATE TABLE yoga.subscription (
  id SERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES yoga.user (id) ON DELETE CASCADE NOT NULL,
  starts DATE NOT NULL,
  ends DATE NOT NULL
);
