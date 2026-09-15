ALTER TABLE media_assets DROP CONSTRAINT IF EXISTS media_assets_dimensions_positive;
ALTER TABLE media_assets
  DROP COLUMN IF EXISTS processed_at,
  DROP COLUMN IF EXISTS failure_code,
  DROP COLUMN IF EXISTS codec_name,
  DROP COLUMN IF EXISTS height,
  DROP COLUMN IF EXISTS width,
  DROP COLUMN IF EXISTS duration_seconds;
