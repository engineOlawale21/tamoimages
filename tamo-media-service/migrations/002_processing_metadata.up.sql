ALTER TABLE media_assets
  ADD COLUMN IF NOT EXISTS duration_seconds double precision,
  ADD COLUMN IF NOT EXISTS width integer,
  ADD COLUMN IF NOT EXISTS height integer,
  ADD COLUMN IF NOT EXISTS codec_name text,
  ADD COLUMN IF NOT EXISTS failure_code text,
  ADD COLUMN IF NOT EXISTS processed_at timestamptz;

ALTER TABLE media_assets
  ADD CONSTRAINT media_assets_dimensions_positive
  CHECK ((width IS NULL OR width > 0) AND (height IS NULL OR height > 0));

COMMENT ON COLUMN media_assets.failure_code IS 'Bounded operational classification only; never store filenames, object keys, command output, or personal data.';
