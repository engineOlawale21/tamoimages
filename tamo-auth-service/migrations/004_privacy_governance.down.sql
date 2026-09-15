ALTER TABLE data_subject_requests
  DROP COLUMN IF EXISTS reviewer_id,
  DROP COLUMN IF EXISTS exemption_code,
  DROP COLUMN IF EXISTS deadline_at,
  DROP COLUMN IF EXISTS fulfilled_at,
  DROP COLUMN IF EXISTS decided_at,
  DROP COLUMN IF EXISTS decision,
  DROP COLUMN IF EXISTS verified_at,
  DROP COLUMN IF EXISTS verification_method;
DROP TABLE IF EXISTS consent_records;
ALTER TABLE users DROP COLUMN IF EXISTS registration_notice_acknowledged_at, DROP COLUMN IF EXISTS registration_notice_version;
DROP TABLE IF EXISTS privacy_notices;
