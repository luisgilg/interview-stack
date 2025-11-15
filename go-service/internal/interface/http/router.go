package http

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/gofiber/swagger"

	"github.com/example/interview-stack/go-service/internal/docs"
)

// RegisterRoutes wires controller endpoints.
func RegisterRoutes(app *fiber.App, controller *ProductController) {
	app.Get("/swagger/*", fiberSwagger.HandlerDefault)
	app.Get("/swagger.json", func(ctx *fiber.Ctx) error {
		return ctx.Type("json").SendString(docs.SwaggerInfo.ReadDoc())
	})

	app.Get("/health", controller.Health)

	app.Get("/products", controller.ListProducts)
	app.Get("/products/:id", controller.GetProduct)
	app.Post("/products", controller.CreateProduct)
	app.Put("/products/:id", controller.UpdateProduct)
	app.Delete("/products/:id", controller.DeleteProduct)
}
