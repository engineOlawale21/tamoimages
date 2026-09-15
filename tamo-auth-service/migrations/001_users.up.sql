CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text NOT NULL,
  first_name text NOT NULL,
  last_name text NOT NULL,
  role text NOT NULL CHECK (role IN ('buyer', 'contributor')),
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT users_email_normalized CHECK (email = lower(email)),
  CONSTRAINT users_email_unique UNIQUE (email)
);

COMMENT ON TABLE users IS 'Identity accounts. Controller purpose: account access and platform security. Retention: active account lifetime, then approved account-deletion schedule subject to legal hold.';
COMMENT ON COLUMN users.email IS 'Authentication and essential account communication; never emitted to operational logs.';
COMMENT ON COLUMN users.password_hash IS 'Adaptive password hash; secret data, never returned or logged.';

CREATE INDEX IF NOT EXISTS users_active_role_idx ON users (role) WHERE deleted_at IS NULL;
