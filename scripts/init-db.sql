-- Seed products
INSERT INTO products (name, description, price, stock, category) VALUES
  ('MacBook Pro 14"',     'Apple M3 Pro, 18GB RAM, 512GB SSD',         1999.99, 45,  'laptops'),
  ('MacBook Air 15"',     'Apple M2, 8GB RAM, 256GB SSD',              1299.99, 80,  'laptops'),
  ('Dell XPS 15',         'Intel i7, 32GB RAM, 1TB NVMe',              1799.99, 30,  'laptops'),
  ('iPhone 15 Pro',       '256GB, Titanium, A17 Pro chip',              999.99, 120, 'phones'),
  ('Samsung Galaxy S24',  '256GB, Snapdragon 8 Gen 3',                  799.99, 90,  'phones'),
  ('Google Pixel 8 Pro',  '256GB, Tensor G3',                           899.99, 60,  'phones'),
  ('AirPods Pro 2',       'Active Noise Cancellation, USB-C',           249.99, 200, 'audio'),
  ('Sony WH-1000XM5',    'Over-ear, ANC, 30hr battery',                349.99, 75,  'audio'),
  ('Bose QC45',           'Over-ear noise cancelling headphones',       279.99, 50,  'audio'),
  ('iPad Pro 12.9"',      'M2 chip, 256GB WiFi',                        999.99, 40,  'tablets'),
  ('iPad Air 5',          'M1 chip, 64GB WiFi',                         599.99, 65,  'tablets'),
  ('Apple Watch Series 9','GPS + Cellular, 45mm, Aluminum',             429.99, 150, 'wearables'),
  ('Garmin Fenix 7',      'Solar, GPS, multisport',                     699.99, 30,  'wearables'),
  ('LG 27" 4K Monitor',   'IPS, 144Hz, USB-C',                         599.99, 25,  'electronics'),
  ('Logitech MX Master 3','Advanced wireless mouse',                     99.99, 300, 'electronics')
ON CONFLICT DO NOTHING;
