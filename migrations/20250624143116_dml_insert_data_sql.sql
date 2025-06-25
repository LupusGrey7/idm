-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
INSERT INTO public.employees(id, name, created_at, updated_at)
VALUES (nextval('public.global_enterprise_sequence'), 'Alice Marcus', NOW(), NOW()),
       (nextval('public.global_enterprise_sequence'), 'Jill Valentine', NOW(), NOW()) ON CONFLICT DO NOTHING;

INSERT INTO public.roles(id, name, created_at, updated_at, employee_id)
VALUES (nextval('public.global_enterprise_sequence'), 'ADMIN', NOW(), NOW(), NULL),
       (nextval('public.global_enterprise_sequence'), 'USER', NOW(), NOW(), NULL) ON CONFLICT DO NOTHING;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
