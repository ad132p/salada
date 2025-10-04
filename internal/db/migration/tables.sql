
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
    category public.category
);


ALTER TABLE public.posts OWNER TO salada;

CREATE TABLE public.users (
    id integer NOT NULL,
    username character varying(255) NOT NULL,
    password text NOT NULL,
    role character varying(50) DEFAULT 'user'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    email text
);


create table public.images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    filepath varchar,
    status varchar(10),
    blog_post_id uuid DEFAULT NULL,
    uploaded_at timestamp
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
