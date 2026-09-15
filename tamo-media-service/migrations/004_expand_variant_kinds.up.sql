ALTER TABLE media_variants DROP CONSTRAINT IF EXISTS media_variants_kind_check;
ALTER TABLE media_variants
  ADD CONSTRAINT media_variants_kind_check
  CHECK (kind IN ('thumbnail', 'poster', 'small', 'medium', 'large'));
