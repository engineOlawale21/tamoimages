CREATE TABLE license_products (
  code text PRIMARY KEY,
  name text NOT NULL,
  media_kind text NOT NULL CHECK (media_kind IN ('image','video','illustration')),
  amount_minor bigint NOT NULL CHECK (amount_minor > 0),
  currency char(3) NOT NULL DEFAULT 'NGN' CHECK (currency = upper(currency))
);

INSERT INTO license_products(code,name,media_kind,amount_minor) VALUES
  ('standard-image','Standard image license','image',1500000),
  ('extended-image','Extended image license','image',4500000),
  ('standard-video','Standard video license','video',3500000),
  ('extended-video','Extended video license','video',9000000),
  ('standard-illustration','Standard illustration license','illustration',1500000),
  ('extended-illustration','Extended illustration license','illustration',4500000);

CREATE TABLE buyer_carts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  buyer_id uuid NOT NULL,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','converted','abandoned')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX buyer_carts_one_active_idx ON buyer_carts(buyer_id) WHERE status='active';

CREATE TABLE buyer_cart_items (
  cart_id uuid NOT NULL REFERENCES buyer_carts(id) ON DELETE CASCADE,
  media_asset_id uuid NOT NULL REFERENCES media_assets(id) ON DELETE RESTRICT,
  license_code text NOT NULL REFERENCES license_products(code) ON DELETE RESTRICT,
  unit_amount_minor bigint NOT NULL CHECK (unit_amount_minor > 0),
  currency char(3) NOT NULL CHECK (currency = upper(currency)),
  added_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(cart_id,media_asset_id)
);

COMMENT ON COLUMN buyer_cart_items.unit_amount_minor IS 'Server-calculated price snapshot in the smallest currency unit; never supplied by a client.';
