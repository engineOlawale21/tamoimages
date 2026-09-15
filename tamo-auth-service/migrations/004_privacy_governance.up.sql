CREATE TABLE privacy_notices (
  version text PRIMARY KEY,
  title text NOT NULL,
  content text NOT NULL,
  effective_at timestamptz NOT NULL,
  retired_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (retired_at IS NULL OR retired_at > effective_at)
);

CREATE UNIQUE INDEX privacy_notices_one_current_idx ON privacy_notices ((retired_at IS NULL)) WHERE retired_at IS NULL;

INSERT INTO privacy_notices(version,title,content,effective_at)
VALUES (
  '2026-09-09',
  'Tamo Images account privacy notice',
  'We use your account identity and contact details to provide and secure your Tamo Images account, communicate essential service information, and meet legal obligations. Identity data stays in the identity service. You may request access, correction, deletion, restriction, objection, portability, consent withdrawal, or make a complaint through your account. Retention depends on the record type and legal holds. Contact the designated privacy owner through the published support channel.',
  '2026-09-09T00:00:00Z'
);

ALTER TABLE users
  ADD COLUMN registration_notice_version text REFERENCES privacy_notices(version) ON DELETE RESTRICT,
  ADD COLUMN registration_notice_acknowledged_at timestamptz;

CREATE TABLE consent_records (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  notice_version text NOT NULL REFERENCES privacy_notices(version) ON DELETE RESTRICT,
  purpose text NOT NULL CHECK (purpose IN ('product_updates','usage_analytics')),
  granted boolean NOT NULL,
  channel text NOT NULL CHECK (channel IN ('account_web','account_api')),
  recorded_at timestamptz NOT NULL DEFAULT now(),
  correlation_id text NOT NULL
);

CREATE INDEX consent_records_user_purpose_idx ON consent_records(user_id,purpose,recorded_at DESC);

ALTER TABLE data_subject_requests
  ADD COLUMN verification_method text NOT NULL DEFAULT 'authenticated_session',
  ADD COLUMN verified_at timestamptz NOT NULL DEFAULT now(),
  ADD COLUMN decision text,
  ADD COLUMN decided_at timestamptz,
  ADD COLUMN fulfilled_at timestamptz,
  ADD COLUMN deadline_at timestamptz NOT NULL DEFAULT (now() + interval '30 days'),
  ADD COLUMN exemption_code text,
  ADD COLUMN reviewer_id uuid REFERENCES users(id) ON DELETE RESTRICT;

COMMENT ON TABLE privacy_notices IS 'Immutable, versioned notices presented at or before applicable collection.';
COMMENT ON TABLE consent_records IS 'Append-only evidence of granular consent choices and withdrawals.';
