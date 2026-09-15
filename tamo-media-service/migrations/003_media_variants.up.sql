CREATE TABLE IF NOT EXISTS media_variants (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  media_asset_id uuid NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('preview', 'poster', 'thumbnail')),
  storage_key text NOT NULL UNIQUE,
  content_type text NOT NULL,
  size_bytes bigint NOT NULL CHECK (size_bytes > 0),
  width integer CHECK (width IS NULL OR width > 0),
  height integer CHECK (height IS NULL OR height > 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (media_asset_id, kind)
);

CREATE INDEX IF NOT EXISTS media_variants_asset_idx ON media_variants (media_asset_id);
COMMENT ON TABLE media_variants IS 'Generated delivery derivatives. Storage keys are private operational data and must not be logged or exposed directly.';
