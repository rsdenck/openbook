package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"openbook/internal/bootstrap"
	"openbook/internal/config"
	"openbook/internal/repository/postgres"
	"openbook/internal/worker"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("OpenBook Ultimate Worker %s\n", version)
		os.Exit(0)
	}

	log.Printf("Starting OpenBook Ultimate Worker %s...", version)

	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Init DB
	db, err := bootstrap.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer db.Close()

	// 3. Init Redis
	rdb, err := bootstrap.InitRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to init Redis: %v", err)
	}
	defer rdb.Close()

	// 4. Repositories
	deploymentRepo := postgres.NewDeploymentRepository(db)
	gitRepo := postgres.NewGitRepository(db)

	// 5. Worker
	w := worker.NewWorker(rdb, deploymentRepo, gitRepo, cfg)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 6. Start
	go w.Start(ctx)

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down worker...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := w.Stop(shutdownCtx); err != nil {
		log.Printf("Worker shutdown error: %v", err)
	}

	log.Println("Worker stopped gracefully")
}
