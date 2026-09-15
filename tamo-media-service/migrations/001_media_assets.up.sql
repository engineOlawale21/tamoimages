CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS media_assets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  contributor_id uuid NOT NULL,
  kind text NOT NULL CHECK (kind IN ('image', 'video', 'illustration')),
  original_filename text NOT NULL,
  storage_key text NOT NULL UNIQUE,
  status text NOT NULL CHECK (status IN ('pending', 'uploaded', 'processing', 'ready', 'failed', 'deleted')),
  content_type text,
  size_bytes bigint CHECK (size_bytes IS NULL OR size_bytes >= 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS media_assets_contributor_created_idx
  ON media_assets (contributor_id, created_at DESC)
  WHERE deleted_at IS NULL;

COMMENT ON TABLE media_assets IS 'Contributor media metadata. Purpose: upload processing, catalogue administration, and licensing preparation. Retention: contributor/account and licensed-record schedule subject to legal hold.';
COMMENT ON COLUMN media_assets.contributor_id IS 'Opaque identity-service subject identifier; do not duplicate contributor profile data here.';
COMMENT ON COLUMN media_assets.original_filename IS 'Contributor supplied metadata; treat as personal data and never include in operational logs.';
COMMENT ON COLUMN media_assets.storage_key IS 'Private object locator; never expose directly or include in operational logs.';
