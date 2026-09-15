CREATE INDEX IF NOT EXISTS media_assets_catalog_search_idx ON media_assets USING gin (
  to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(description, '') || ' ' || coalesce(location, ''))
) WHERE deleted_at IS NULL AND status = 'ready';

CREATE INDEX IF NOT EXISTS media_assets_catalog_keywords_idx ON media_assets USING gin (keywords)
  WHERE deleted_at IS NULL AND status = 'ready';

CREATE INDEX IF NOT EXISTS media_assets_catalog_filters_idx
  ON media_assets (kind, usage_type, created_at DESC)
  WHERE deleted_at IS NULL AND status = 'ready';

CREATE INDEX IF NOT EXISTS media_batches_approved_idx
  ON media_batches (reviewed_at DESC, id)
  WHERE deleted_at IS NULL AND status = 'approved';
