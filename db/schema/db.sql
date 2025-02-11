--
-- PostgreSQL database dump
--

-- Dumped from database version 17.2 (Debian 17.2-1.pgdg120+1)
-- Dumped by pg_dump version 17.2 (Debian 17.2-1.pgdg120+1)

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: account; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.account (
    account_id text NOT NULL,
    name text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    status_data jsonb,
    repo text,
    repo_status text DEFAULT 'inactive'::text NOT NULL,
    repo_status_data jsonb,
    secret text NOT NULL,
    data jsonb,
    resource_commit_hash text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.account OWNER TO postgres;

--
-- Name: resource_key_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.resource_key_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.resource_key_seq OWNER TO postgres;

--
-- Name: resource; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.resource (
    account_id text DEFAULT current_setting('app.account_id'::text) NOT NULL,
    resource_key bigint DEFAULT nextval('public.resource_key_seq'::regclass) NOT NULL,
    resource_id uuid NOT NULL,
    name text NOT NULL,
    version text,
    description text,
    status text DEFAULT 'new'::text NOT NULL,
    status_data jsonb,
    key_field text NOT NULL,
    key_regex text,
    clear_condition text,
    clear_after bigint DEFAULT (((60 * 60) * 24) * 30) NOT NULL,
    clear_delay bigint DEFAULT 0 NOT NULL,
    data jsonb,
    source text,
    commit_hash text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by bigint,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_by bigint
);


ALTER TABLE public.resource OWNER TO postgres;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO postgres;

--
-- Name: tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tag (
    account_id text DEFAULT current_setting('app.account_id'::text) NOT NULL,
    tag_key text NOT NULL,
    tag_val text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    status_data jsonb,
    data jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by bigint,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_by bigint
);


ALTER TABLE public.tag OWNER TO postgres;

--
-- Name: tag_obj_key_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tag_obj_key_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.tag_obj_key_seq OWNER TO postgres;

--
-- Name: tag_obj; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tag_obj (
    account_id text DEFAULT current_setting('app.account_id'::text) NOT NULL,
    tag_obj_key bigint DEFAULT nextval('public.tag_obj_key_seq'::regclass) NOT NULL,
    tag_type text NOT NULL,
    tag_obj_id text NOT NULL,
    tag_key text NOT NULL,
    tag_val text,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by bigint,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_by bigint
);


ALTER TABLE public.tag_obj OWNER TO postgres;

--
-- Name: user_key_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_key_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_key_seq OWNER TO postgres;

--
-- Name: user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."user" (
    user_key bigint DEFAULT nextval('public.user_key_seq'::regclass) NOT NULL,
    user_id text NOT NULL,
    password text,
    email text,
    last_name text,
    first_name text,
    status text DEFAULT 'active'::text NOT NULL,
    data jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by bigint,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_by bigint
);


ALTER TABLE public."user" OWNER TO postgres;

--
-- Name: account account_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT account_pkey PRIMARY KEY (account_id);


--
-- Name: resource resource_account_id_resource_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT resource_account_id_resource_id_key UNIQUE (account_id, resource_id);


--
-- Name: resource resource_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT resource_pkey PRIMARY KEY (account_id, resource_key);


--
-- Name: resource resource_resource_key_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT resource_resource_key_key UNIQUE (resource_key);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: tag_obj tag_obj_account_id_tag_type_tag_obj_id_tag_key_tag_val_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag_obj
    ADD CONSTRAINT tag_obj_account_id_tag_type_tag_obj_id_tag_key_tag_val_key UNIQUE (account_id, tag_type, tag_obj_id, tag_key, tag_val);


--
-- Name: tag_obj tag_obj_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag_obj
    ADD CONSTRAINT tag_obj_pkey PRIMARY KEY (account_id, tag_obj_key);


--
-- Name: tag_obj tag_obj_tag_obj_key_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag_obj
    ADD CONSTRAINT tag_obj_tag_obj_key_key UNIQUE (tag_obj_key);


--
-- Name: tag tag_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_pkey PRIMARY KEY (account_id, tag_key, tag_val);


--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (user_key);


--
-- Name: user user_user_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_user_id_key UNIQUE (user_id);


--
-- Name: resource resource_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT resource_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(account_id) ON DELETE CASCADE;


--
-- Name: resource resource_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT resource_created_by_fkey FOREIGN KEY (created_by) REFERENCES public."user"(user_key) ON DELETE SET NULL;


--
-- Name: resource resource_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT resource_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public."user"(user_key) ON DELETE SET NULL;


--
-- Name: tag tag_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(account_id) ON DELETE CASCADE;


--
-- Name: tag tag_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_created_by_fkey FOREIGN KEY (created_by) REFERENCES public."user"(user_key) ON DELETE SET NULL;


--
-- Name: tag_obj tag_obj_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag_obj
    ADD CONSTRAINT tag_obj_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(account_id) ON DELETE CASCADE;


--
-- Name: tag_obj tag_obj_account_id_tag_key_tag_val_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag_obj
    ADD CONSTRAINT tag_obj_account_id_tag_key_tag_val_fkey FOREIGN KEY (account_id, tag_key, tag_val) REFERENCES public.tag(account_id, tag_key, tag_val) ON DELETE CASCADE;


--
-- Name: tag_obj tag_obj_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag_obj
    ADD CONSTRAINT tag_obj_created_by_fkey FOREIGN KEY (created_by) REFERENCES public."user"(user_key) ON DELETE SET NULL;


--
-- Name: tag_obj tag_obj_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag_obj
    ADD CONSTRAINT tag_obj_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public."user"(user_key) ON DELETE SET NULL;


--
-- Name: tag tag_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public."user"(user_key) ON DELETE SET NULL;


--
-- Name: account; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.account ENABLE ROW LEVEL SECURITY;

--
-- Name: account account_isolation_policy; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY account_isolation_policy ON public.account USING (((current_setting('app.account_id'::text) = 'sys'::text) OR (account_id = current_setting('app.account_id'::text))));


--
-- Name: resource account_isolation_policy; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY account_isolation_policy ON public.resource USING ((account_id = current_setting('app.account_id'::text)));


--
-- Name: tag account_isolation_policy; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY account_isolation_policy ON public.tag USING ((account_id = current_setting('app.account_id'::text)));


--
-- Name: tag_obj account_isolation_policy; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY account_isolation_policy ON public.tag_obj USING ((account_id = current_setting('app.account_id'::text)));


--
-- Name: resource; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.resource ENABLE ROW LEVEL SECURITY;

--
-- Name: tag; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.tag ENABLE ROW LEVEL SECURITY;

--
-- Name: tag_obj; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.tag_obj ENABLE ROW LEVEL SECURITY;

--
-- Name: TABLE account; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.account TO "api-db-user";


--
-- Name: SEQUENCE resource_key_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.resource_key_seq TO "api-db-user";


--
-- Name: TABLE resource; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.resource TO "api-db-user";


--
-- Name: TABLE schema_migrations; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.schema_migrations TO "api-db-user";


--
-- Name: TABLE tag; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.tag TO "api-db-user";


--
-- Name: SEQUENCE tag_obj_key_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.tag_obj_key_seq TO "api-db-user";


--
-- Name: TABLE tag_obj; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.tag_obj TO "api-db-user";


--
-- Name: SEQUENCE user_key_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.user_key_seq TO "api-db-user";


--
-- Name: TABLE "user"; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public."user" TO "api-db-user";


--
-- PostgreSQL database dump complete
--

