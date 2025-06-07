-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS public;
CREATE SEQUENCE public.global_enterprise_sequence
    START WITH 10
    INCREMENT BY 1
    MINVALUE 1
    NO MAXVALUE
    CACHE 1
    CYCLE;

CREATE TABLE IF NOT EXISTS public.employees (
    id BIGINT PRIMARY KEY DEFAULT nextval('public.global_enterprise_sequence'),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS public.roles (
    id BIGINT PRIMARY KEY DEFAULT nextval('public.global_enterprise_sequence'),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    employee_id BIGINT DEFAULT NULL,
    CONSTRAINT roles_name_unique UNIQUE (name),
    CONSTRAINT fk_employee FOREIGN KEY (employee_id) REFERENCES public.employees(id) ON DELETE CASCADE
    );

CREATE INDEX roles_employee_id_idx ON public.roles (employee_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.roles CASCADE;
DROP TABLE IF EXISTS public.employees CASCADE;
DROP SEQUENCE IF EXISTS global_enterprise_sequence;
-- +goose StatementEnd
