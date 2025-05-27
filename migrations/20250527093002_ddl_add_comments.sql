-- +goose Up
-- +goose StatementBegin
-- Комментарии для таблицы employees
COMMENT ON TABLE public.employees IS 'Employee Information';
COMMENT ON COLUMN public.employees.id IS 'Уникальный идентификатор сотрудника';
COMMENT ON COLUMN public.employees.name IS 'ФИО сотрудника';
COMMENT ON COLUMN public.employees.created_at IS 'Дата создания записи';
COMMENT ON COLUMN public.employees.updated_at IS 'Дата последнего обновления записи';

-- Комментарии для таблицы roles
COMMENT ON TABLE public.roles IS 'Информация о ролях';
COMMENT ON COLUMN public.roles.id IS 'Уникальный идентификатор роли';
COMMENT ON COLUMN public.roles.name IS 'Наименование роли';
COMMENT ON COLUMN public.roles.created_at IS 'Дата создания записи';
COMMENT ON COLUMN public.roles.updated_at IS 'Дата последнего обновления записи';
COMMENT ON COLUMN public.roles.employee_id IS 'Ссылка на сотрудника (FK)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
