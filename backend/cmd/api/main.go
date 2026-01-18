package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/routes"
	"github.com/barsukov/quiz-sprint/backend/pkg/database"

	_ "github.com/barsukov/quiz-sprint/backend/docs"
)

// @title Quiz Sprint API
// @version 1.0
// @description Quiz Sprint TMA Backend API - A Telegram Mini App for interactive quizzes
// @termsOfService https://quiz-sprint-tma.online/terms

// @contact.name API Support
// @contact.url https://quiz-sprint-tma.online/support
// @contact.email support@quiz-sprint-tma.online

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api/v1

// @schemes http https

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// ========================================
	// Database Connection
	// ========================================
	var db *sql.DB
	dbConfig := database.LoadConfigFromEnv()

	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Printf("⚠️  Failed to connect to PostgreSQL: %v", err)
		log.Println("⚠️  User endpoints will not be available without database")
		db = nil
	}

	// Ensure database connection is closed on shutdown
	if db != nil {
		defer func() {
			if err := db.Close(); err != nil {
				log.Printf("Error closing database connection: %v", err)
			}
		}()
	}

	// ========================================
	// Fiber App Setup
	// ========================================
	app := fiber.New(fiber.Config{
		AppName:      "Quiz Sprint API",
		ServerHeader: "Quiz Sprint",
		ErrorHandler: errorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	// CORS configuration
	corsOrigins := getEnv("CORS_ORIGINS", "*")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: corsOrigins != "*", // Only allow credentials if not wildcard
	}))

	// Health check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "quiz-sprint-api",
		})
	})

	// Setup routes (pass database connection)
	routes.SetupRoutes(app, db)

	// ========================================
	// Start Server
	// ========================================
	port := getEnv("PORT", "3000")
	log.Printf("🚀 Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func errorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    code,
			"message": message,
		},
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
