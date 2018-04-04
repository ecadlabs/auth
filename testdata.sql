--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: users; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE users (
    id uuid NOT NULL,
    email text,
    password_hash text NOT NULL,
    first_name text,
    last_name text,
    added timestamp without time zone DEFAULT now(),
    modified timestamp without time zone DEFAULT now()
);


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO users VALUES ('0d4ef66e-8017-4c17-9b50-3a4ff263f6e8', 'e.asphyx@gmail.com', '$2a$10$jHWtDuzh4ZgVGOrorJ2F5uXSn5g9JlR40iAOwJtVRpLvlpqp5BnDq', 'Eugene', 'Zagidullin', '2018-03-29 02:06:10.849697', '2018-03-29 02:06:10.849697');
INSERT INTO users VALUES ('14153e5c-c593-42f5-b606-1aabbdbc6f54', 'jev@ecadlabs.com', '$2a$10$C5qDedxsAJUa85DQEEor2.eqcSrjS3J5gmQNkS4X65hc1iOFgd4de', 'Jev', 'Björsell', '2018-03-29 02:09:26.333747', '2018-03-29 02:09:26.333747');


--
-- Name: users_email_key; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--

