-- +goose Up
-- +goose StatementBegin

ALTER TABLE roles
    ADD COLUMN IF NOT EXISTS company_id INTEGER REFERENCES companies(id) ON DELETE SET NULL;

ALTER TABLE roles
    DROP CONSTRAINT IF EXISTS roles_name_key,
    DROP CONSTRAINT IF EXISTS roles_slug_key;

DROP INDEX IF EXISTS idx_roles_name_company_id_unique;
DROP INDEX IF EXISTS idx_roles_slug_company_id_unique;
DROP INDEX IF EXISTS idx_roles_global_name_unique;
DROP INDEX IF EXISTS idx_roles_global_slug_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_name_company_id_unique
    ON roles (name, company_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_slug_company_id_unique
    ON roles (slug, company_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_global_name_unique
    ON roles (name)
    WHERE company_id IS NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_global_slug_unique
    ON roles (slug)
    WHERE company_id IS NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_roles_company_id ON roles(company_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_roles_company_id;
DROP INDEX IF EXISTS idx_roles_global_slug_unique;
DROP INDEX IF EXISTS idx_roles_global_name_unique;
DROP INDEX IF EXISTS idx_roles_slug_company_id_unique;
DROP INDEX IF EXISTS idx_roles_name_company_id_unique;

ALTER TABLE roles
    ADD CONSTRAINT roles_name_key UNIQUE (name),
    ADD CONSTRAINT roles_slug_key UNIQUE (slug);

ALTER TABLE roles
    DROP COLUMN IF EXISTS company_id;

-- +goose StatementEnd
