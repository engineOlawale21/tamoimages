CREATE TABLE buyer_collections (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  buyer_id uuid NOT NULL,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (buyer_id,name)
);
CREATE INDEX buyer_collections_owner_idx ON buyer_collections(buyer_id,created_at DESC);

CREATE TABLE buyer_collection_items (
  collection_id uuid NOT NULL REFERENCES buyer_collections(id) ON DELETE CASCADE,
  media_asset_id uuid NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
  added_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(collection_id,media_asset_id)
);

CREATE TABLE buyer_favourites (
  buyer_id uuid NOT NULL,
  media_asset_id uuid NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
  added_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(buyer_id,media_asset_id)
);
CREATE INDEX buyer_favourites_owner_idx ON buyer_favourites(buyer_id,added_at DESC);

COMMENT ON TABLE buyer_collections IS 'Buyer-owned catalogue organisation using opaque identity subject identifiers.';
COMMENT ON TABLE buyer_collection_items IS 'Minimal relationship between a buyer collection and an approved catalogue asset.';
COMMENT ON TABLE buyer_favourites IS 'Minimal buyer-to-approved-media favourite relationship.';
