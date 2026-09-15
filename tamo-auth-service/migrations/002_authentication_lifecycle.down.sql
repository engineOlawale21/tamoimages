DROP TABLE IF EXISTS security_audit_events;
DROP TABLE IF EXISTS contributor_profiles;
DROP TABLE IF EXISTS buyer_profiles;
DROP TABLE IF EXISTS refresh_sessions;
DROP TABLE IF EXISTS password_recovery_tokens;
DROP TABLE IF EXISTS email_verification_tokens;

ALTER TABLE users
  DROP COLUMN IF EXISTS last_login_at,
  DROP COLUMN IF EXISTS password_changed_at,
  DROP COLUMN IF EXISTS email_verified_at,
  DROP COLUMN IF EXISTS status;
