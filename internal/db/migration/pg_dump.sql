--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5 (Debian 17.5-1.pgdg120+1)
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
-- Name: category; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.category AS ENUM (
    'cs',
    'politics',
    'plants',
    'cine',
    'lit'
);


ALTER TYPE public.category OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: http_sessions; Type: TABLE; Schema: public; Owner: salada
--

CREATE TABLE public.http_sessions (
    id bigint NOT NULL,
    key bytea,
    data bytea,
    created_on timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    modified_on timestamp with time zone,
    expires_on timestamp with time zone
);


ALTER TABLE public.http_sessions OWNER TO salada;

--
-- Name: http_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: salada
--

CREATE SEQUENCE public.http_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.http_sessions_id_seq OWNER TO salada;

--
-- Name: http_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: salada
--

ALTER SEQUENCE public.http_sessions_id_seq OWNED BY public.http_sessions.id;


--
-- Name: posts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.posts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    content text NOT NULL,
    author_id uuid,
    author_name character varying(255),
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    tags text,
    category public.category
);


ALTER TABLE public.posts OWNER TO postgres;

--
-- Name: sessions; Type: TABLE; Schema: public; Owner: salada
--

CREATE TABLE public.sessions (
    id text NOT NULL,
    data bytea,
    created_on timestamp with time zone DEFAULT now(),
    expires_on timestamp with time zone
);


ALTER TABLE public.sessions OWNER TO salada;

--
-- Name: users; Type: TABLE; Schema: public; Owner: salada
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username character varying(255) NOT NULL,
    password text NOT NULL,
    role character varying(50) DEFAULT 'user'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    email text
);


ALTER TABLE public.users OWNER TO salada;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: salada
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO salada;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: salada
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: http_sessions id; Type: DEFAULT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.http_sessions ALTER COLUMN id SET DEFAULT nextval('public.http_sessions_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: http_sessions http_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.http_sessions
    ADD CONSTRAINT http_sessions_pkey PRIMARY KEY (id);


--
-- Name: posts posts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id);


--
-- Name: posts posts_slug_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_slug_key UNIQUE (slug);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: http_sessions_expiry_idx; Type: INDEX; Schema: public; Owner: salada
--

CREATE INDEX http_sessions_expiry_idx ON public.http_sessions USING btree (expires_on);


--
-- Name: http_sessions_key_idx; Type: INDEX; Schema: public; Owner: salada
--

CREATE INDEX http_sessions_key_idx ON public.http_sessions USING btree (key);


--
-- Name: idx_sessions_expires_on; Type: INDEX; Schema: public; Owner: salada
--

CREATE INDEX idx_sessions_expires_on ON public.sessions USING btree (expires_on);


--
-- PostgreSQL database dump complete
--

