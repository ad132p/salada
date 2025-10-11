
CREATE TYPE public.category AS ENUM (
    'cs',
    'politics',
    'plants',
    'cine',
    'lit',
    'football',
    'random'
);

CREATE TABLE public.posts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(255) NOT NULL,
    slug character varying(255) UNIQUE NOT NULL,
    content text NOT NULL,
    author_id uuid,
    author_name character varying(255),
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    tags text[],
    category public.category
);

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
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

