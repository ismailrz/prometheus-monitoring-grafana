package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	appmetrics "github.com/prometheus-stack/backend/pkg/metrics"
)

type Order struct {
	ID            string      `json:"id"`
	CustomerEmail string      `json:"customer_email"`
	Status        string      `json:"status"`
	TotalAmount   float64     `json:"total_amount"`
	Items         []OrderItem `json:"items,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        string  `json:"id"`
	OrderID   string  `json:"order_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type CreateOrderRequest struct {
	CustomerEmail string            `json:"customer_email" binding:"required,email"`
	Items         []OrderItemInput  `json:"items"          binding:"required,min=1,dive"`
}

type OrderItemInput struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity"   binding:"required,min=1"`
}

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) List(ctx context.Context, status string) ([]Order, error) {
	t := appmetrics.NewDBTimer("select", "orders")
	defer t.ObserveDuration()

	var (
		rows *sql.Rows
		err  error
	)
	if status != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, customer_email, status, total_amount, created_at, updated_at
			 FROM orders WHERE status=$1 ORDER BY created_at DESC`, status)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, customer_email, status, total_amount, created_at, updated_at
			 FROM orders ORDER BY created_at DESC LIMIT 100`)
	}
	if err != nil {
		appmetrics.DbQueryErrors.WithLabelValues("select", "orders").Inc()
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.CustomerEmail, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*Order, error) {
	t := appmetrics.NewDBTimer("select", "orders")
	defer t.ObserveDuration()

	var o Order
	err := r.db.QueryRowContext(ctx,
		`SELECT id, customer_email, status, total_amount, created_at, updated_at
		 FROM orders WHERE id=$1`, id).
		Scan(&o.ID, &o.CustomerEmail, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.getItems(ctx, id)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return &o, nil
}

func (r *OrderRepository) Create(ctx context.Context, req CreateOrderRequest) (*Order, error) {
	t := appmetrics.NewDBTimer("insert", "orders")
	defer t.ObserveDuration()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	var o Order
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO orders (customer_email, status) VALUES ($1, 'pending')
		 RETURNING id, customer_email, status, total_amount, created_at, updated_at`,
		req.CustomerEmail).
		Scan(&o.ID, &o.CustomerEmail, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return nil, err
	}

	var total float64
	for _, item := range req.Items {
		var price float64
		if err := tx.QueryRowContext(ctx,
			`SELECT price FROM products WHERE id=$1`, item.ProductID).
			Scan(&price); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				appmetrics.BusinessErrors.WithLabelValues("product_not_found").Inc()
				return nil, errors.New("product not found: " + item.ProductID)
			}
			return nil, err
		}

		var oi OrderItem
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity, unit_price)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, order_id, product_id, quantity, unit_price`,
			o.ID, item.ProductID, item.Quantity, price).
			Scan(&oi.ID, &oi.OrderID, &oi.ProductID, &oi.Quantity, &oi.UnitPrice); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, oi)
		total += price * float64(item.Quantity)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE orders SET total_amount=$1, updated_at=NOW() WHERE id=$2`, total, o.ID); err != nil {
		return nil, err
	}
	o.TotalAmount = total

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	appmetrics.OrdersCreatedTotal.Inc()
	appmetrics.OrderValueTotal.Add(total)
	return &o, nil
}

func (r *OrderRepository) getItems(ctx context.Context, orderID string) ([]OrderItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, product_id, quantity, unit_price FROM order_items WHERE order_id=$1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var i OrderItem
		if err := rows.Scan(&i.ID, &i.OrderID, &i.ProductID, &i.Quantity, &i.UnitPrice); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
