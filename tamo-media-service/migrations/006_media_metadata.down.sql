ALTER TABLE media_assets
  DROP CONSTRAINT IF EXISTS media_assets_keywords_count,
  DROP CONSTRAINT IF EXISTS media_assets_usage_type,
  DROP CONSTRAINT IF EXISTS media_assets_location_length,
  DROP CONSTRAINT IF EXISTS media_assets_description_length,
  DROP CONSTRAINT IF EXISTS media_assets_title_length;
ALTER TABLE media_assets DROP COLUMN IF EXISTS usage_type, DROP COLUMN IF EXISTS location,
  DROP COLUMN IF EXISTS keywords, DROP COLUMN IF EXISTS description, DROP COLUMN IF EXISTS title;
