CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE microsaas.roles (
  role_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  display_name VARCHAR(100) UNIQUE NOT NULL,
  appointment_quota int
);

CREATE TABLE microsaas.users (
  user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email VARCHAR(255) UNIQUE NOT NULL,
  pass_hash VARCHAR(255),
  first_name VARCHAR(255),
  last_name VARCHAR(255),
  role_id UUID references microsaas.roles(role_id),
  enabled BOOLEAN DEFAULT true,
  status VARCHAR(100),
  source VARCHAR(100) NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
  deleted_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE TABLE microsaas.accounts (
  user_id UUID PRIMARY KEY references microsaas.users(user_id),
  failed_login_counter INTEGER,
  failed_login_lock TIMESTAMP WITHOUT TIME ZONE,
  last_failed_login TIMESTAMP WITHOUT TIME ZONE,
  confirmation_token VARCHAR(100),
  confirmation_ttl TIMESTAMP WITHOUT TIME ZONE,
  confirmed BOOLEAN,
  recovery_token VARCHAR(100),
  recovery_ttl TIMESTAMP WITHOUT TIME ZONE,
  last_recovery TIMESTAMP WITHOUT TIME ZONE,
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
  deleted_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX idx_accounts_confirmation_token ON microsaas.accounts(confirmation_token);
CREATE INDEX idx_accounts_recovery_token ON microsaas.accounts(recovery_token);
