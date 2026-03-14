--
-- PostgreSQL database dump
--

\restrict FU1normnuAkibep5Ea1rETD6g4i0qOuWfsbzhiumRQ6x8xNrieydfCimIhZaWPt

-- Dumped from database version 18.0 (Debian 18.0-1.pgdg13+3)
-- Dumped by pg_dump version 18.0 (Debian 18.0-1.pgdg13+3)

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
-- Name: category; Type: TYPE; Schema: public; Owner: salada
--

CREATE TYPE public.category AS ENUM (
    'cs',
    'politics',
    'plants',
    'cine',
    'lit',
    'football',
    'random'
);


ALTER TYPE public.category OWNER TO salada;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: access_logs; Type: TABLE; Schema: public; Owner: salada
--

CREATE TABLE public.access_logs (
    id integer NOT NULL,
    "timestamp" timestamp with time zone DEFAULT now(),
    client_ip character varying(45) NOT NULL,
    method character varying(10) NOT NULL,
    path character varying(255) NOT NULL,
    status_code integer NOT NULL,
    latency_ms integer NOT NULL
);


ALTER TABLE public.access_logs OWNER TO salada;

--
-- Name: access_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: salada
--

CREATE SEQUENCE public.access_logs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.access_logs_id_seq OWNER TO salada;

--
-- Name: access_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: salada
--

ALTER SEQUENCE public.access_logs_id_seq OWNED BY public.access_logs.id;


--
-- Name: comments; Type: TABLE; Schema: public; Owner: salada
--

CREATE TABLE public.comments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    content text NOT NULL,
    blog_post_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    author_name character varying(255)
);


ALTER TABLE public.comments OWNER TO salada;

--
-- Name: images; Type: TABLE; Schema: public; Owner: salada
--

CREATE TABLE public.images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    filepath character varying,
    status character varying(10),
    blog_post_id uuid,
    uploaded_at timestamp without time zone
);


ALTER TABLE public.images OWNER TO salada;

--
-- Name: posts; Type: TABLE; Schema: public; Owner: salada
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
    tags text[],
    category public.category,
    seen integer DEFAULT 0 NOT NULL,
    likes integer DEFAULT 0,
    thumbnail_position text DEFAULT 'center'::text
);


ALTER TABLE public.posts OWNER TO salada;

--
-- Name: users; Type: TABLE; Schema: public; Owner: salada
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username character varying(255) NOT NULL,
    password text NOT NULL,
    role character varying(50) DEFAULT 'user'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    email text
);


ALTER TABLE public.users OWNER TO salada;

--
-- Name: access_logs id; Type: DEFAULT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.access_logs ALTER COLUMN id SET DEFAULT nextval('public.access_logs_id_seq'::regclass);


--
-- Name: access_logs access_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.access_logs
    ADD CONSTRAINT access_logs_pkey PRIMARY KEY (id);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: posts posts_pkey; Type: CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id);


--
-- Name: posts posts_slug_key; Type: CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_slug_key UNIQUE (slug);


--
-- Name: users users_uniq; Type: CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_uniq UNIQUE (username);


--
-- Name: posts fk_author_name; Type: FK CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT fk_author_name FOREIGN KEY (author_name) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: comments fk_blog_post; Type: FK CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT fk_blog_post FOREIGN KEY (blog_post_id) REFERENCES public.posts(id) ON DELETE CASCADE;


--
-- Name: images fk_blog_post_id; Type: FK CONSTRAINT; Schema: public; Owner: salada
--

ALTER TABLE ONLY public.images
    ADD CONSTRAINT fk_blog_post_id FOREIGN KEY (blog_post_id) REFERENCES public.posts(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict FU1normnuAkibep5Ea1rETD6g4i0qOuWfsbzhiumRQ6x8xNrieydfCimIhZaWPt


--
-- Name: ws_metrics; Type: TABLE; Schema: public; Owner: salada
--

CREATE TABLE public.ws_metrics (
    id SERIAL PRIMARY KEY,
    "timestamp" timestamp with time zone DEFAULT now(),
    client_ip character varying(45) NOT NULL,
    path character varying(255) NOT NULL,
    bytes_read bigint NOT NULL DEFAULT 0,
    bytes_written bigint NOT NULL DEFAULT 0,
    duration_ms bigint NOT NULL DEFAULT 0,
    streamer character varying(255) DEFAULT 'unknown'::character varying NOT NULL,
    watcher character varying(255) DEFAULT 'anon'::character varying NOT NULL
);

ALTER TABLE public.ws_metrics OWNER TO salada;
