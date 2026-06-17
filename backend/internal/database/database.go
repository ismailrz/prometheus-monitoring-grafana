package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	appmetrics "github.com/prometheus-stack/backend/pkg/metrics"
)

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}

// CollectStats pushes db pool stats to Prometheus gauges periodically.
func CollectStats(db *sql.DB) {
	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for range t.C {
			s := db.Stats()
			appmetrics.DbConnectionsOpen.Set(float64(s.OpenConnections))
			appmetrics.DbConnectionsInUse.Set(float64(s.InUse))
			appmetrics.DbConnectionsIdle.Set(float64(s.Idle))
		}
	}()
}

const schema = `
CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    price       NUMERIC(10,2) NOT NULL,
    stock       INTEGER       NOT NULL DEFAULT 0,
    category    VARCHAR(100)  NOT NULL DEFAULT 'general',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_email VARCHAR(255)  NOT NULL,
    status         VARCHAR(50)   NOT NULL DEFAULT 'pending',
    total_amount   NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID          NOT NULL REFERENCES orders(id)   ON DELETE CASCADE,
    product_id UUID          NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity   INTEGER       NOT NULL,
    unit_price NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_orders_status     ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_email      ON orders(customer_email);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
`
