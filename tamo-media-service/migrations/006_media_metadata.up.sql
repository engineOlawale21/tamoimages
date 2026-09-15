ALTER TABLE media_assets
  ADD COLUMN IF NOT EXISTS title text,
  ADD COLUMN IF NOT EXISTS description text,
  ADD COLUMN IF NOT EXISTS keywords text[] NOT NULL DEFAULT '{}',
  ADD COLUMN IF NOT EXISTS location text,
  ADD COLUMN IF NOT EXISTS usage_type text;

ALTER TABLE media_assets
  ADD CONSTRAINT media_assets_title_length CHECK (title IS NULL OR char_length(title) BETWEEN 1 AND 160),
  ADD CONSTRAINT media_assets_description_length CHECK (description IS NULL OR char_length(description) <= 2000),
  ADD CONSTRAINT media_assets_location_length CHECK (location IS NULL OR char_length(location) <= 200),
  ADD CONSTRAINT media_assets_usage_type CHECK (usage_type IS NULL OR usage_type IN ('creative', 'editorial')),
  ADD CONSTRAINT media_assets_keywords_count CHECK (cardinality(keywords) <= 50);

COMMENT ON COLUMN media_assets.location IS 'Contributor-supplied general capture location; avoid precise private addresses.';
