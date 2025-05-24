--comment: create-schema 20250524_lupusgrey
CREATE SCHEMA IF NOT EXISTS public;

--comment: Drop schema and table objects
DROP TABLE IF EXISTS public.roles CASCADE;
DROP TABLE IF EXISTS public.employees CASCADE;

--comment: Create hibernate_sequence
CREATE DOMAIN enterprise_id AS BIGINT
    DEFAULT nextval('global_enterprise_sequence')
    NOT NULL
    CHECK (VALUE > 0);

--comment: Создаем глобальную sequence с настройками
CREATE SEQUENCE global_enterprise_sequence
    START WITH 10
    INCREMENT BY 1
    MINVALUE 1
    NO MAXVALUE
    CACHE 1
    CYCLE;

--comment: create-employees-table
CREATE TABLE IF NOT EXISTS public.employees (
    id enterprise_id PRIMARY KEY,
    name VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--comment: create-roles-table
CREATE TABLE IF NOT EXISTS public.roles (
    id enterprise_id PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    employee_id BIGINT NOT NULL,
    CONSTRAINT fk_employee FOREIGN KEY (employee_id) REFERENCES public.employees(id) ON DELETE CASCADE
);

--comment: create-index-for-role
CREATE INDEX roles_employee_id_idx ON public.roles (employee_id);

--comment: create-comment-for-employees
COMMENT ON TABLE public.employees IS 'Employee Information';
COMMENT ON TABLE public.employees.id IS 'Unique identifier for each employee record';
COMMENT ON TABLE public.employees.name IS 'date when first was created record';
COMMENT ON TABLE public.employees.create_at IS 'date when record was updated';
COMMENT ON TABLE public.employees.update_is IS 'first name Employee';

--comment: create-comment-for-roles
COMMENT ON TABLE public.roles IS 'Role Information';
COMMENT ON TABLE public.roles.id IS 'Unique identifier for each role record';
COMMENT ON TABLE public.roles.name IS 'date when first was created record';
COMMENT ON TABLE public.roles.create_at IS 'date when record was updated';
COMMENT ON TABLE public.roles.update_is IS 'ROLE in System';
COMMENT ON COLUMN public.roles.employee_id IS 'Reference Key to employee record';