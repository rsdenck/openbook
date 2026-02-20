package main

import (
	"fmt"
	"log"
	"os"

	"openbook/internal/bootstrap"
	"openbook/internal/config"
	"openbook/internal/handler"
	"openbook/internal/middleware"
	"openbook/internal/repository/postgres"
	"openbook/internal/service"
	"openbook/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("OpenBook Ultimate API %s\n", version)
		os.Exit(0)
	}

	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Database
	db, err := bootstrap.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer db.Close()

	// 3. Redis
	rdb, err := bootstrap.InitRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to init Redis: %v", err)
	}
	defer rdb.Close()

	// 4. Repositories
	deploymentRepo := postgres.NewDeploymentRepository(db)
	gitRepo := postgres.NewGitRepository(db)
	// siteRepo := postgres.NewSiteRepository(db)
	// domainRepo := postgres.NewDomainRepository(db)
	auditRepo := postgres.NewAuditLogRepository(db)

	// 5. Services
	publisher := service.NewPublisher(rdb)

	// 6. UseCases
	deploymentUC := usecase.NewDeploymentUseCase(deploymentRepo, auditRepo, publisher)
	gitUC := usecase.NewGitUseCase(gitRepo)

	// 7. Handlers
	deploymentHandler := handler.NewDeploymentHandler(deploymentUC)
	branchHandler := handler.NewBranchHandler(gitUC)
	mergeHandler := handler.NewMergeHandler(gitUC)

	// 8. Fiber App
	app := fiber.New()
	app.Use(logger.New())

	// Health Check
	app.Get("/health", func(c *fiber.Ctx) error {
		// Basic check for DB/Redis connection
		if err := db.Ping(); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("DB Down")
		}
		if err := rdb.Ping(c.Context()).Err(); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Redis Down")
		}
		return c.SendString("OK")
	})

	// API Group
	api := app.Group("/api/v1")
	api.Use(middleware.AuthMiddleware) // JWT + Workspace ID

	// Git Engine Routes
	api.Post("/branches", branchHandler.Create)
	api.Get("/branches", branchHandler.Get)
	api.Post("/merge", mergeHandler.Merge)

	// Deployment Routes
	api.Post("/deployments", deploymentHandler.Create)
	api.Get("/deployments/:id", deploymentHandler.GetByID)

	// Start Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
