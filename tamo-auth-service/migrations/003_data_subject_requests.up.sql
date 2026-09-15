CREATE TABLE IF NOT EXISTS data_subject_requests (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  kind text NOT NULL CHECK (kind IN ('access','correction','portability','deletion','restriction','objection','consent_withdrawal','complaint')),
  status text NOT NULL DEFAULT 'accepted' CHECK (status IN ('accepted','identity_review','in_progress','completed','rejected')),
  correlation_id text NOT NULL,
  received_at timestamptz NOT NULL DEFAULT now(),
  completed_at timestamptz,
  legal_hold_until timestamptz,
  CHECK (completed_at IS NULL OR completed_at >= received_at)
);

CREATE INDEX IF NOT EXISTS data_subject_requests_user_idx ON data_subject_requests(user_id, received_at DESC);
CREATE INDEX IF NOT EXISTS data_subject_requests_open_idx ON data_subject_requests(received_at) WHERE completed_at IS NULL;
COMMENT ON TABLE data_subject_requests IS 'NDPA/NDPR rights-request workflow records; request bodies and identity documents must not be stored here.';
