CREATE TABLE media_deletion_evidence (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  media_asset_id uuid NOT NULL,
  contributor_id uuid NOT NULL,
  reason text NOT NULL CHECK (reason IN ('subject_request','contributor_request','retention_expiry')),
  object_count integer NOT NULL CHECK (object_count >= 0),
  target_hash text NOT NULL CHECK (char_length(target_hash) = 64),
  completed_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX media_deletion_evidence_asset_idx ON media_deletion_evidence(media_asset_id,completed_at);
COMMENT ON TABLE media_deletion_evidence IS 'Append-only, privacy-minimised evidence that media database and object-storage deletion completed.';
