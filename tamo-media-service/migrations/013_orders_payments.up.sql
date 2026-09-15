CREATE TABLE commerce_orders (
  id uuid PRIMARY KEY,
  buyer_id uuid NOT NULL,
  idempotency_key text NOT NULL,
  payment_reference text NOT NULL UNIQUE,
  status text NOT NULL DEFAULT 'pending_payment' CHECK (status IN ('pending_payment','paid','cancelled','payment_failed')),
  currency char(3) NOT NULL,
  total_amount_minor bigint NOT NULL CHECK (total_amount_minor > 0),
  checkout_url text,
  paid_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(buyer_id,idempotency_key)
);

CREATE TABLE commerce_order_items (
  order_id uuid NOT NULL REFERENCES commerce_orders(id) ON DELETE RESTRICT,
  media_asset_id uuid NOT NULL REFERENCES media_assets(id) ON DELETE RESTRICT,
  contributor_id uuid NOT NULL,
  license_code text NOT NULL,
  license_name text NOT NULL,
  unit_amount_minor bigint NOT NULL CHECK (unit_amount_minor > 0),
  currency char(3) NOT NULL,
  PRIMARY KEY(order_id,media_asset_id)
);

CREATE TABLE payment_webhook_events (
  provider text NOT NULL,
  event_key text NOT NULL,
  payment_reference text NOT NULL,
  received_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(provider,event_key)
);

COMMENT ON TABLE commerce_order_items IS 'Immutable license and price snapshots copied from a server-priced cart.';
COMMENT ON TABLE payment_webhook_events IS 'Minimal webhook deduplication evidence; raw payment payloads are not retained.';
