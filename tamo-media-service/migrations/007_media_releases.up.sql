CREATE TABLE IF NOT EXISTS media_releases (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  media_asset_id uuid NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
  contributor_id uuid NOT NULL,
  release_type text NOT NULL CHECK (release_type IN ('model', 'property')),
  filename text NOT NULL CHECK (char_length(filename) BETWEEN 1 AND 255),
  storage_key text NOT NULL UNIQUE,
  content_type text NOT NULL CHECK (content_type IN ('application/pdf', 'image/jpeg', 'image/png')),
  size_bytes bigint NOT NULL CHECK (size_bytes > 0),
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'uploaded')),
  upload_expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS media_releases_asset_idx ON media_releases(media_asset_id, created_at);
CREATE INDEX IF NOT EXISTS media_releases_pending_expiry_idx ON media_releases(upload_expires_at) WHERE status='pending';
COMMENT ON TABLE media_releases IS 'Private model/property release documents; maximum three per media asset is enforced transactionally by the repository.';
COMMENT ON COLUMN media_releases.storage_key IS 'Private object locator; never expose directly or log.';
