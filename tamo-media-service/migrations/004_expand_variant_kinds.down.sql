DELETE FROM media_variants WHERE kind IN ('small', 'medium', 'large');
ALTER TABLE media_variants DROP CONSTRAINT IF EXISTS media_variants_kind_check;
ALTER TABLE media_variants
  ADD CONSTRAINT media_variants_kind_check
  CHECK (kind IN ('preview', 'poster', 'thumbnail'));
