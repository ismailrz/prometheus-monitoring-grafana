package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus-stack/backend/internal/config"
	"github.com/prometheus-stack/backend/internal/database"
	"github.com/prometheus-stack/backend/internal/handlers"
	"github.com/prometheus-stack/backend/internal/middleware"
	"github.com/prometheus-stack/backend/internal/models"
	appmetrics "github.com/prometheus-stack/backend/pkg/metrics"
)

func main() {
	cfg := config.Load()
	appmetrics.Init(cfg.Version, cfg.Env)

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		fmt.Fprintf(os.Stderr, "database migrate: %v\n", err)
		os.Exit(1)
	}

	// Collect DB pool stats every 15s into Prometheus gauges.
	database.CollectStats(db)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Metrics())
	r.Use(gin.Recovery())

	// Observability
	r.GET("/health", handlers.Health(db))
	r.GET("/ready", handlers.Ready(db))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1
	v1 := r.Group("/api/v1")
	{
		productRepo := models.NewProductRepository(db)
		ph := handlers.NewProductHandler(productRepo)
		v1.GET("/products", ph.List)
		v1.GET("/products/:id", ph.Get)
		v1.POST("/products", ph.Create)
		v1.PUT("/products/:id", ph.Update)
		v1.DELETE("/products/:id", ph.Delete)

		orderRepo := models.NewOrderRepository(db)
		oh := handlers.NewOrderHandler(orderRepo)
		v1.GET("/orders", oh.List)
		v1.GET("/orders/:id", oh.Get)
		v1.POST("/orders", oh.Create)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("server listening on :%s (env=%s version=%s)\n", cfg.Port, cfg.Env, cfg.Version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server: %v\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown: %v\n", err)
	}
}
