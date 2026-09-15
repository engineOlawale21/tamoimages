ALTER TABLE users
  ADD COLUMN IF NOT EXISTS status text
    CHECK (status IN ('pending_verification', 'active', 'suspended', 'deletion_requested')),
  ADD COLUMN IF NOT EXISTS email_verified_at timestamptz,
  ADD COLUMN IF NOT EXISTS password_changed_at timestamptz NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS last_login_at timestamptz;

-- Accounts created before verification existed were already permitted to sign in.
UPDATE users
SET status = 'active', email_verified_at = COALESCE(email_verified_at, created_at)
WHERE status IS NULL;

ALTER TABLE users ALTER COLUMN status SET DEFAULT 'pending_verification';
ALTER TABLE users ALTER COLUMN status SET NOT NULL;

CREATE TABLE IF NOT EXISTS email_verification_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash bytea NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  consumed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (expires_at > created_at)
);

CREATE TABLE IF NOT EXISTS password_recovery_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash bytea NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  consumed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (expires_at > created_at)
);

CREATE TABLE IF NOT EXISTS refresh_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  family_id uuid NOT NULL,
  token_hash bytea NOT NULL UNIQUE,
  user_agent_hash bytea,
  ip_prefix_hash bytea,
  issued_at timestamptz NOT NULL DEFAULT now(),
  last_used_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  rotated_at timestamptz,
  revoked_at timestamptz,
  revocation_reason text,
  replaced_by_session_id uuid REFERENCES refresh_sessions(id) ON DELETE SET NULL,
  CHECK (expires_at > issued_at)
);

CREATE TABLE IF NOT EXISTS buyer_profiles (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  display_name text CHECK (char_length(display_name) <= 100),
  organisation text CHECK (char_length(organisation) <= 160),
  country_code char(2),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS contributor_profiles (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  display_name text CHECK (char_length(display_name) <= 100),
  profession text CHECK (char_length(profession) <= 100),
  biography text CHECK (char_length(biography) <= 1000),
  country_code char(2),
  city text CHECK (char_length(city) <= 100),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS security_audit_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  event_type text NOT NULL,
  outcome text NOT NULL CHECK (outcome IN ('success', 'failure')),
  correlation_id text,
  subject_hash bytea,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS email_verification_tokens_expiry_idx ON email_verification_tokens (expires_at) WHERE consumed_at IS NULL;
CREATE INDEX IF NOT EXISTS password_recovery_tokens_expiry_idx ON password_recovery_tokens (expires_at) WHERE consumed_at IS NULL;
CREATE INDEX IF NOT EXISTS refresh_sessions_user_active_idx ON refresh_sessions (user_id, expires_at) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS refresh_sessions_family_idx ON refresh_sessions (family_id);
CREATE INDEX IF NOT EXISTS security_audit_events_retention_idx ON security_audit_events (occurred_at);

COMMENT ON TABLE email_verification_tokens IS 'Hashed single-use email verification credentials. Default retention: expiry plus 24 hours.';
COMMENT ON TABLE password_recovery_tokens IS 'Hashed single-use password recovery credentials. Default retention: expiry plus 24 hours.';
COMMENT ON TABLE refresh_sessions IS 'Durable session and rotation-family records; opaque credentials are never stored.';
COMMENT ON TABLE security_audit_events IS 'Privacy-minimised authentication security events. Never store credentials or raw request payloads.';
