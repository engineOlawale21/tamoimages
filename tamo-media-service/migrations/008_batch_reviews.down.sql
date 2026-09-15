ALTER TABLE media_batches DROP CONSTRAINT IF EXISTS media_batches_review_feedback_length;
ALTER TABLE media_batches DROP COLUMN IF EXISTS reviewed_at, DROP COLUMN IF EXISTS review_feedback;
