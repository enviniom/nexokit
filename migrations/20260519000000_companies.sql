-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS companies (
    id SERIAL PRIMARY KEY,
    public_id CHAR(26) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(120) NOT NULL UNIQUE,
    domain VARCHAR(255),
    subdomain VARCHAR(120),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by INTEGER,
    updated_by INTEGER
);

CREATE INDEX IF NOT EXISTS idx_companies_deleted_at ON companies(deleted_at);
CREATE INDEX IF NOT EXISTS idx_companies_domain ON companies(domain);
CREATE INDEX IF NOT EXISTS idx_companies_subdomain ON companies(subdomain);
CREATE INDEX IF NOT EXISTS idx_companies_status ON companies(status);
CREATE INDEX IF NOT EXISTS idx_companies_created_by ON companies(created_by);
CREATE INDEX IF NOT EXISTS idx_companies_updated_by ON companies(updated_by);

ALTER TABLE users
    ADD CONSTRAINT fk_users_company_id
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_users_company_id ON users(company_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_users_company_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_company_id;
DROP TABLE IF EXISTS companies;

-- +goose StatementEnd
