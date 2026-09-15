ALTER TABLE media_batches
  ADD COLUMN IF NOT EXISTS review_feedback text,
  ADD COLUMN IF NOT EXISTS reviewed_at timestamptz;

ALTER TABLE media_batches
  ADD CONSTRAINT media_batches_review_feedback_length
  CHECK (review_feedback IS NULL OR char_length(review_feedback) <= 2000);

COMMENT ON COLUMN media_batches.review_feedback IS 'Bounded reviewer feedback visible to the contributor; do not include private reviewer notes.';
