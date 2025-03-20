-- Create keyspace with SimpleStrategy for single-node development
CREATE KEYSPACE IF NOT EXISTS order_service
WITH REPLICATION = {
  'class': 'SimpleStrategy',
  'replication_factor': 1
};

-- Use the order_service keyspace
USE order_service;

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
  id UUID,
  user_id TEXT,
  items BLOB, -- Serialized JSON of order items
  total_amount DOUBLE,
  status TEXT,
  shipping_address BLOB, -- Serialized JSON of address
  billing_address BLOB, -- Serialized JSON of address
  payment_id TEXT,
  shipping_id TEXT,
  notes TEXT,
  promotion_codes SET<TEXT>,
  discounts BLOB, -- Serialized JSON of discounts
  tax_amount DOUBLE,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  completed_at TIMESTAMP,
  cancelled_at TIMESTAMP,
  version INT,
  PRIMARY KEY (id)
);

-- Create secondary index for user_id queries
CREATE INDEX IF NOT EXISTS orders_user_id_idx ON orders (user_id);

-- Create secondary index for status queries
CREATE INDEX IF NOT EXISTS orders_status_idx ON orders (status);

-- Table for accessing orders by user (efficient querying)
CREATE TABLE IF NOT EXISTS orders_by_user (
  user_id TEXT,
  order_id UUID,
  status TEXT,
  total_amount DOUBLE,
  created_at TIMESTAMP,
  PRIMARY KEY (user_id, order_id)
);

-- Table for accessing orders by status (efficient querying)
CREATE TABLE IF NOT EXISTS orders_by_status (
  status TEXT,
  order_id UUID,
  user_id TEXT,
  created_at TIMESTAMP,
  PRIMARY KEY (status, created_at, order_id)
) WITH CLUSTERING ORDER BY (created_at DESC, order_id ASC);

-- Order events table for event sourcing
CREATE TABLE IF NOT EXISTS order_events (
  id UUID,
  order_id UUID,
  type TEXT,
  data BLOB, -- Serialized JSON of event data
  version INT,
  timestamp TIMESTAMP,
  user_id TEXT,
  PRIMARY KEY (order_id, version, id)
) WITH CLUSTERING ORDER BY (version ASC, id ASC);

-- Create secondary index for order event type queries
CREATE INDEX IF NOT EXISTS order_events_type_idx ON order_events (type);

-- Payments table
CREATE TABLE IF NOT EXISTS payments (
  id UUID,
  order_id UUID,
  amount DOUBLE,
  currency TEXT,
  method TEXT,
  status TEXT,
  transaction_id TEXT,
  gateway_response TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  completed_at TIMESTAMP,
  failed_at TIMESTAMP,
  PRIMARY KEY (id)
);

-- Create secondary index for order_id queries in payments
CREATE INDEX IF NOT EXISTS payments_order_id_idx ON payments (order_id);

-- Shipping table
CREATE TABLE IF NOT EXISTS shipping (
  id UUID,
  order_id UUID,
  carrier TEXT,
  tracking_number TEXT,
  status TEXT,
  estimated_delivery TIMESTAMP,
  shipped_at TIMESTAMP,
  delivered_at TIMESTAMP,
  shipping_method TEXT,
  shipping_cost DOUBLE,
  notes TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  PRIMARY KEY (id)
);

-- Create secondary index for order_id queries in shipping
CREATE INDEX IF NOT EXISTS shipping_order_id_idx ON shipping (order_id);

-- Create secondary index for tracking_number queries in shipping
CREATE INDEX IF NOT EXISTS shipping_tracking_number_idx ON shipping (tracking_number);