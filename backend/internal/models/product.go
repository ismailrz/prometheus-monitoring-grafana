package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	appmetrics "github.com/prometheus-stack/backend/pkg/metrics"
)

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name        string  `json:"name"        binding:"required,min=1,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price"       binding:"required,gt=0"`
	Stock       int     `json:"stock"       binding:"min=0"`
	Category    string  `json:"category"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"        binding:"required,min=1,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price"       binding:"required,gt=0"`
	Stock       int     `json:"stock"       binding:"min=0"`
	Category    string  `json:"category"`
}

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List(ctx context.Context, category string) ([]Product, error) {
	t := appmetrics.NewDBTimer("select", "products")
	defer t.ObserveDuration()

	var (
		rows *sql.Rows
		err  error
	)

	if category != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, name, description, price, stock, category, created_at, updated_at
			 FROM products WHERE category = $1 ORDER BY created_at DESC`, category)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, name, description, price, stock, category, created_at, updated_at
			 FROM products ORDER BY created_at DESC`)
	}
	if err != nil {
		appmetrics.DbQueryErrors.WithLabelValues("select", "products").Inc()
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*Product, error) {
	t := appmetrics.NewDBTimer("select", "products")
	defer t.ObserveDuration()

	var p Product
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, price, stock, category, created_at, updated_at
		 FROM products WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		appmetrics.DbQueryErrors.WithLabelValues("select", "products").Inc()
	}
	return &p, err
}

func (r *ProductRepository) Create(ctx context.Context, req CreateProductRequest) (*Product, error) {
	t := appmetrics.NewDBTimer("insert", "products")
	defer t.ObserveDuration()

	category := req.Category
	if category == "" {
		category = "general"
	}

	var p Product
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO products (name, description, price, stock, category)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, name, description, price, stock, category, created_at, updated_at`,
		req.Name, req.Description, req.Price, req.Stock, category).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		appmetrics.DbQueryErrors.WithLabelValues("insert", "products").Inc()
		return nil, err
	}
	appmetrics.ProductsCreatedTotal.Inc()
	return &p, nil
}

func (r *ProductRepository) Update(ctx context.Context, id string, req UpdateProductRequest) (*Product, error) {
	t := appmetrics.NewDBTimer("update", "products")
	defer t.ObserveDuration()

	var p Product
	err := r.db.QueryRowContext(ctx,
		`UPDATE products
		 SET name=$1, description=$2, price=$3, stock=$4, category=$5, updated_at=NOW()
		 WHERE id=$6
		 RETURNING id, name, description, price, stock, category, created_at, updated_at`,
		req.Name, req.Description, req.Price, req.Stock, req.Category, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		appmetrics.DbQueryErrors.WithLabelValues("update", "products").Inc()
	}
	return &p, err
}

func (r *ProductRepository) Delete(ctx context.Context, id string) (bool, error) {
	t := appmetrics.NewDBTimer("delete", "products")
	defer t.ObserveDuration()

	res, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		appmetrics.DbQueryErrors.WithLabelValues("delete", "products").Inc()
		return false, err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		appmetrics.ProductsDeletedTotal.Inc()
	}
	return n > 0, nil
}
