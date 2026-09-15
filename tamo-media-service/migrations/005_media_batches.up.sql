CREATE TABLE IF NOT EXISTS media_batches (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  contributor_id uuid NOT NULL,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'in_review', 'approved', 'rejected')),
  submitted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS media_batches_contributor_created_idx
  ON media_batches (contributor_id, created_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS media_batch_items (
  batch_id uuid NOT NULL REFERENCES media_batches(id) ON DELETE CASCADE,
  media_asset_id uuid NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
  added_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (batch_id, media_asset_id),
  UNIQUE (media_asset_id)
);

COMMENT ON TABLE media_batches IS 'Contributor-owned workflow containers used to submit media for review.';
COMMENT ON TABLE media_batch_items IS 'Assignment of a media asset to one contributor review batch.';
