package main

import (
	"log"
	"os"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/caarlos0/env/v6"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/italia/publiccode-validator-api/internal/common"
	"github.com/italia/publiccode-validator-api/internal/handlers"
	"github.com/italia/publiccode-validator-api/internal/jsondecoder"
)

func main() {
	app := Setup()
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}

func Setup() *fiber.App {
	if err := env.Parse(&common.EnvironmentConfig); err != nil {
		panic(err)
	}

	methods := append(fiber.DefaultMethods, "QUERY") //nolint:gocritic // We want a new slice here

	app := fiber.New(fiber.Config{
		ErrorHandler: common.CustomErrorHandler,
		// Fiber doesn't set DisallowUnknownFields by default
		// (https://github.com/gofiber/fiber/issues/2601)
		JSONDecoder:    jsondecoder.UnmarshalDisallowUnknownFields,
		RequestMethods: methods,
	})

	// Automatically recover panics in handlers
	app.Use(recover.New())

	app.Use(cache.New(cache.Config{
		Next: func(ctx *fiber.Ctx) bool {
			// Don't cache /status
			return ctx.Route().Path == "/v1/status"
		},
		Methods:      []string{fiber.MethodGet, fiber.MethodHead},
		CacheControl: true,
		Expiration:   10 * time.Second, //nolint:gomnd
		KeyGenerator: func(ctx *fiber.Ctx) string {
			return ctx.Path() + string(ctx.Context().QueryArgs().QueryString())
		},
	}))

	prometheus := fiberprometheus.New(os.Args[0])
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	setupHandlers(app)

	return app
}

func setupHandlers(app *fiber.App) {
	validateHandler := handlers.NewPubliccodeymlValidatorHandler()

	v1 := app.Group("/v1")

	v1.Add("QUERY", "/validate", validateHandler.Query)
}
