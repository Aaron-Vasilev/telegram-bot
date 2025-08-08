--
-- PostgreSQL database dump
--

-- Dumped from database version 16.2 (Debian 16.2-1.pgdg120+2)
-- Dumped by pg_dump version 17.5

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: yoga; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA yoga;


ALTER SCHEMA yoga OWNER TO postgres;

--
-- Name: get_lesson(date); Type: FUNCTION; Schema: yoga; Owner: postgres
--

CREATE FUNCTION yoga.get_lesson(date_param date) RETURNS json
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN (
    SELECT
      json_build_object(
        'lesson_id', yoga.lesson.id,
        'date', yoga.lesson.date,
        'registered', (
          SELECT
            json_agg(
              json_build_object(
                'user_id', yoga.user.id,
                'name', yoga.user.name
              )
            )
          FROM
            yoga.registered_users
            JOIN yoga.user ON yoga.user.id = ANY (yoga.registered_users.registered)
          WHERE
            yoga.registered_users.lesson_id = yoga.lesson.id
          )
        )
    FROM
      yoga.lesson
    WHERE
      yoga.lesson.date = date_param
  );
END;
$$;


ALTER FUNCTION yoga.get_lesson(date_param date) OWNER TO postgres;

--
-- Name: get_lesson(date, time without time zone); Type: FUNCTION; Schema: yoga; Owner: postgres
--

CREATE FUNCTION yoga.get_lesson(date_param date, time_param time without time zone) RETURNS json
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN (
    SELECT
      json_build_object(
        'lessonId', yoga.lesson.id,
        'date', yoga.lesson.date,
        'time', yoga.lesson.time,
		'max', yoga.lesson.max,
        'description', yoga.lesson.description,
        'registered', COALESCE((
            SELECT
              json_agg(
                json_build_object(
                  'id', yoga.user.id,
                  'name', yoga.user.name,
                  'username', yoga.user.username,
				  'emoji', yoga.user.emoji
                )
              )
            FROM
              yoga.registered_users
              JOIN yoga.user ON yoga.user.id = ANY (yoga.registered_users.registered)
            WHERE
              yoga.registered_users.lesson_id = yoga.lesson.id
          ),
          json_build_array()
          )
        )
    FROM
      yoga.lesson
    WHERE
      yoga.lesson.date = date_param AND
      yoga.lesson.time = time_param
  );
END;
$$;


ALTER FUNCTION yoga.get_lesson(date_param date, time_param time without time zone) OWNER TO postgres;

--
-- Name: attendance_id_seq; Type: SEQUENCE; Schema: yoga; Owner: postgres
--

CREATE SEQUENCE yoga.attendance_id_seq
    START WITH 14
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE yoga.attendance_id_seq OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: attendance; Type: TABLE; Schema: yoga; Owner: postgres
--

CREATE TABLE yoga.attendance (
    id integer DEFAULT nextval('yoga.attendance_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    lesson_id integer NOT NULL,
    date date NOT NULL
);


ALTER TABLE yoga.attendance OWNER TO postgres;

--
-- Name: lesson; Type: TABLE; Schema: yoga; Owner: postgres
--

CREATE TABLE yoga.lesson (
    id integer NOT NULL,
    date date NOT NULL,
    "time" time without time zone NOT NULL,
    description text NOT NULL,
    max integer DEFAULT 5 NOT NULL,
    poll_id bigint
);


ALTER TABLE yoga.lesson OWNER TO postgres;

--
-- Name: available_lessons; Type: VIEW; Schema: yoga; Owner: postgres
--

CREATE VIEW yoga.available_lessons AS
 SELECT id,
    date,
    "time",
    description,
    max
   FROM yoga.lesson
  WHERE (date >= (now())::date);


ALTER VIEW yoga.available_lessons OWNER TO postgres;

--
-- Name: lesson_id_seq; Type: SEQUENCE; Schema: yoga; Owner: postgres
--

CREATE SEQUENCE yoga.lesson_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE yoga.lesson_id_seq OWNER TO postgres;

--
-- Name: lesson_id_seq; Type: SEQUENCE OWNED BY; Schema: yoga; Owner: postgres
--

ALTER SEQUENCE yoga.lesson_id_seq OWNED BY yoga.lesson.id;


--
-- Name: membership; Type: TABLE; Schema: yoga; Owner: postgres
--

CREATE TABLE yoga.membership (
    user_id bigint NOT NULL,
    starts date NOT NULL,
    ends date NOT NULL,
    type integer DEFAULT 0 NOT NULL,
    lessons_avaliable integer DEFAULT 0 NOT NULL
);


ALTER TABLE yoga.membership OWNER TO postgres;

--
-- Name: migrations; Type: TABLE; Schema: yoga; Owner: postgres
--

CREATE TABLE yoga.migrations (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    run_on timestamp without time zone NOT NULL
);


ALTER TABLE yoga.migrations OWNER TO postgres;

--
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: yoga; Owner: postgres
--

CREATE SEQUENCE yoga.migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE yoga.migrations_id_seq OWNER TO postgres;

--
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: yoga; Owner: postgres
--

ALTER SEQUENCE yoga.migrations_id_seq OWNED BY yoga.migrations.id;


--
-- Name: registered_users; Type: TABLE; Schema: yoga; Owner: postgres
--

CREATE TABLE yoga.registered_users (
    lesson_id integer NOT NULL,
    registered bigint[]
);


ALTER TABLE yoga.registered_users OWNER TO postgres;

--
-- Name: registered_users_lesson_id_seq; Type: SEQUENCE; Schema: yoga; Owner: postgres
--

CREATE SEQUENCE yoga.registered_users_lesson_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE yoga.registered_users_lesson_id_seq OWNER TO postgres;

--
-- Name: registered_users_lesson_id_seq; Type: SEQUENCE OWNED BY; Schema: yoga; Owner: postgres
--

ALTER SEQUENCE yoga.registered_users_lesson_id_seq OWNED BY yoga.registered_users.lesson_id;


--
-- Name: subscription; Type: TABLE; Schema: yoga; Owner: postgres
--

CREATE TABLE yoga.subscription (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    starts date NOT NULL,
    ends date NOT NULL
);


ALTER TABLE yoga.subscription OWNER TO postgres;

--
-- Name: subscription_id_seq; Type: SEQUENCE; Schema: yoga; Owner: postgres
--

CREATE SEQUENCE yoga.subscription_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE yoga.subscription_id_seq OWNER TO postgres;

--
-- Name: subscription_id_seq; Type: SEQUENCE OWNED BY; Schema: yoga; Owner: postgres
--

ALTER SEQUENCE yoga.subscription_id_seq OWNED BY yoga.subscription.id;


--
-- Name: token; Type: TABLE; Schema: yoga; Owner: postgres
--

CREATE TABLE yoga.token (
    id uuid,
    type integer,
    created date,
    valid boolean,
    user_id bigint
);


ALTER TABLE yoga.token OWNER TO postgres;

--
-- Name: user; Type: TABLE; Schema: yoga; Owner: postgres
--

CREATE TABLE yoga."user" (
    id bigint NOT NULL,
    username character varying(32),
    name character varying(128) NOT NULL,
    emoji character varying(32) DEFAULT '🧘🏿'::character varying NOT NULL,
    is_blocked boolean DEFAULT false
);


ALTER TABLE yoga."user" OWNER TO postgres;

--
-- Name: user_id_seq; Type: SEQUENCE; Schema: yoga; Owner: postgres
--

CREATE SEQUENCE yoga.user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE yoga.user_id_seq OWNER TO postgres;

--
-- Name: user_id_seq; Type: SEQUENCE OWNED BY; Schema: yoga; Owner: postgres
--

ALTER SEQUENCE yoga.user_id_seq OWNED BY yoga."user".id;


--
-- Name: lesson id; Type: DEFAULT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.lesson ALTER COLUMN id SET DEFAULT nextval('yoga.lesson_id_seq'::regclass);


--
-- Name: migrations id; Type: DEFAULT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.migrations ALTER COLUMN id SET DEFAULT nextval('yoga.migrations_id_seq'::regclass);


--
-- Name: registered_users lesson_id; Type: DEFAULT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.registered_users ALTER COLUMN lesson_id SET DEFAULT nextval('yoga.registered_users_lesson_id_seq'::regclass);


--
-- Name: subscription id; Type: DEFAULT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.subscription ALTER COLUMN id SET DEFAULT nextval('yoga.subscription_id_seq'::regclass);


--
-- Name: user id; Type: DEFAULT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga."user" ALTER COLUMN id SET DEFAULT nextval('yoga.user_id_seq'::regclass);


--
-- Name: attendance attendance_pkey; Type: CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.attendance
    ADD CONSTRAINT attendance_pkey PRIMARY KEY (id);


--
-- Name: lesson lesson_pkey; Type: CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.lesson
    ADD CONSTRAINT lesson_pkey PRIMARY KEY (id);


--
-- Name: membership membership_pkey; Type: CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.membership
    ADD CONSTRAINT membership_pkey PRIMARY KEY (user_id);


--
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);


--
-- Name: registered_users registered_users_pkey; Type: CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.registered_users
    ADD CONSTRAINT registered_users_pkey PRIMARY KEY (lesson_id);


--
-- Name: subscription subscription_pkey; Type: CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.subscription
    ADD CONSTRAINT subscription_pkey PRIMARY KEY (id);


--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (id);


--
-- Name: user user_username_key; Type: CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga."user"
    ADD CONSTRAINT user_username_key UNIQUE (username);


--
-- Name: lesson_date_idx; Type: INDEX; Schema: yoga; Owner: postgres
--

CREATE INDEX lesson_date_idx ON yoga.lesson USING btree (date);


--
-- Name: attendance lesson_id_foreign_keysd; Type: FK CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.attendance
    ADD CONSTRAINT lesson_id_foreign_keysd FOREIGN KEY (lesson_id) REFERENCES yoga.lesson(id);


--
-- Name: membership membership_user_id_fkey; Type: FK CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.membership
    ADD CONSTRAINT membership_user_id_fkey FOREIGN KEY (user_id) REFERENCES yoga."user"(id) ON DELETE CASCADE;


--
-- Name: registered_users registered_users_lesson_id_fkey; Type: FK CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.registered_users
    ADD CONSTRAINT registered_users_lesson_id_fkey FOREIGN KEY (lesson_id) REFERENCES yoga.lesson(id) ON DELETE CASCADE;


--
-- Name: subscription subscription_user_id_fkey; Type: FK CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.subscription
    ADD CONSTRAINT subscription_user_id_fkey FOREIGN KEY (user_id) REFERENCES yoga."user"(id) ON DELETE CASCADE;


--
-- Name: attendance user_id_foreign_keysd; Type: FK CONSTRAINT; Schema: yoga; Owner: postgres
--

ALTER TABLE ONLY yoga.attendance
    ADD CONSTRAINT user_id_foreign_keysd FOREIGN KEY (user_id) REFERENCES yoga."user"(id);

