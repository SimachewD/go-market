package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-market/config"
	"go-market/internal/api"
	"go-market/internal/jobs"
	"go-market/internal/models"
	"go-market/internal/repo/cache"
	"go-market/internal/repo/postgres"
)

func main() {
    // Load config
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // Connect to Postgres
    db, err := postgres.NewPostgresDB(cfg.PostgresDSN)
    if err != nil {
        log.Fatalf("failed to connect to DB: %v", err)
    }
    
    // Run migrations
    if err := db.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{}); err != nil {
        log.Fatalf("failed to migrate database: %v", err)
    }

    defer postgres.CloseDB(db)

    // Connect to Redis
    redisClient := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
    defer redisClient.Close()

    // wsManager := ws.NewManager() // create websocket manager

    // start workers
    jobQueue := jobs.NewJobQueue(db, 3) // 3 worker goroutines
    jobQueue.Start()
    // defer jobQueue.Stop()

    // Setup router
    r := api.NewRouter(db, redisClient, cfg.JWTSecret, jobQueue)

    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.Port),
        Handler: r,
    }

    // Run server in a goroutine
    go func() {
        log.Printf("Starting server on port %d...", cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    jobs_ctx, jobs_cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer jobs_cancel()
    // Stop accepting new jobs
    jobQueue.Stop(jobs_ctx)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exiting")
}
