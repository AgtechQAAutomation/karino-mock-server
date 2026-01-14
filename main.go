package main

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/shyamsundaar/karino-mock-server/controllers"
	"github.com/shyamsundaar/karino-mock-server/initializers"
	"github.com/shyamsundaar/karino-mock-server/middleware"

	_ "github.com/shyamsundaar/karino-mock-server/docs"
)

// @title ERP Farmer & Sales Order Integration API
// @version 1.0
// @description This is a sample CRUD API for managing Notes and Farmer Details
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@example.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name APIKey

// @security ApiKeyAuth
// @host localhost:8001
// @BasePath /
func main() {
	app := fiber.New()

	// 1. Path Normalization Middleware
	// This captures // and replaces it with / so the router doesn't 404
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.Contains(path, "//") {
			newPath := strings.ReplaceAll(path, "//", "/")
			return c.Redirect(newPath, fiber.StatusMovedPermanently)
		}
		return c.Next()
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, APIKey",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	// Swagger Route
	app.Get("/swagger/*", swagger.HandlerDefault)

	micro := fiber.New()
	app.Mount("/", micro)

	micro.Route("/spic_to_erp", func(router fiber.Router) {
		router.Use(middleware.ApiKeyAuth)
		router.Use(middleware.JSONProviderMiddleware)

		router.Route("/customers", func(router fiber.Router) {

			// Grouping by coopId to keep it clean
			router.Route("/:coopId", func(cust fiber.Router) {

				// Farmer Routes
				cust.Post("/farmers", controllers.CreateCustomerDetailHandler)
				cust.Get("/farmers", controllers.FindCustomerDetailsHandler)
				cust.Get("/farmers/:farmerId", controllers.GetCustomerDetailHandler)

				// Sales Orders Group
				cust.Route("/salesorders", func(sales fiber.Router) {
					// STATIC ROUTES FIRST
					// Matches: /salesorders/deliverydocuments
					sales.Post("/deliverydocuments", controllers.CreateCustomerDeliveryDocumentDetailsHandler)
					sales.Get("/deliverydocuments", controllers.GetCustomerDeliveryDocumentDetailHandler)

					// PARAMETRIC ROUTES SECOND
					// Matches: /salesorders/:orderId
					sales.Get("/:orderId", controllers.GetCustomerSalesOrderDetailsHandler)
					// Matches: /salesorders/:orderId/deliverydocuments
					sales.Get("/:orderId/deliverydocuments", controllers.GetDeliveryDetailParticularHandler)

					// Base Sales Order Routes
					sales.Post("/", controllers.CreateCustomerSalesOrderHandler)
					sales.Get("/", controllers.GetCustomerSalesDetailHandler)
				})

				// Delivery Proof Routes
				cust.Post("/deliverydocuments/:deliveryNoteId/proof", controllers.CreateDeliveryDocumentsProofHandler)
				cust.Get("/deliverydocuments/invoices", controllers.GetDeliveryDocumentsProofHandler)
				cust.Get("/deliverydocuments/:deliveryNoteId/invoices", controllers.GetDeliveryDocumentsProofParticularHandler)
			})
		})

		router.Route("/vendors", func(router fiber.Router) {
			router.Route("/:coopId", func(vend fiber.Router) {
				vend.Post("/farmers", controllers.CreateVendorDetailHandler)
				vend.Get("/farmers", controllers.FindVendorDetailsHandler)
				vend.Get("/farmers/:farmerId", controllers.GetVendorDetailHandler)
			})
		})
	})

	log.Fatal(app.Listen(":8001"))
}

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatalln("Failed to load environment variables! \n", err.Error())
	}
	initializers.ConnectDB(&config)
}
