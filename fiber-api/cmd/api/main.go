// Command api levanta el servicio HTTP de procesamiento de matrices.
package main

import (
	"log"
	"time"

	"fiber-api/internal/client"
	"fiber-api/internal/config"
	"fiber-api/internal/handler"
	"fiber-api/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.Load()

	app := fiber.New(fiber.Config{
		AppName:      "matrix-factorization-api",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Composición de dependencias: se construyen de adentro hacia afuera.
	statsClient := client.NewStatisticsClient(cfg.StatisticsAPIURL, cfg.HTTPTimeout)
	processor := service.NewMatrixProcessor(statsClient)
	matrixHandler := handler.NewMatrixHandler(processor)

	v1 := app.Group("/api/v1")
	matrixHandler.Register(v1)

	log.Printf("servicio escuchando en :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("no se pudo iniciar el servidor: %v", err)
	}
}
