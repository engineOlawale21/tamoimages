CREATE TABLE download_entitlements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id uuid NOT NULL REFERENCES commerce_orders(id) ON DELETE RESTRICT,
  buyer_id uuid NOT NULL,
  media_asset_id uuid NOT NULL REFERENCES media_assets(id) ON DELETE RESTRICT,
  license_code text NOT NULL,
  granted_at timestamptz NOT NULL DEFAULT now(),
  revoked_at timestamptz,
  UNIQUE(order_id,media_asset_id)
);
CREATE INDEX download_entitlements_buyer_idx ON download_entitlements(buyer_id,granted_at DESC);

CREATE TABLE licensed_download_audit (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  entitlement_id uuid NOT NULL REFERENCES download_entitlements(id) ON DELETE RESTRICT,
  correlation_id text NOT NULL,
  requested_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE download_entitlements IS 'Durable proof that a paid order grants a buyer access to a licensed original.';
COMMENT ON TABLE licensed_download_audit IS 'Minimal download audit evidence; signed URLs and storage keys are never retained.';
