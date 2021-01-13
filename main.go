package main

import (
	"CryptGuard_Back-end/controllers"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	usersAPI := app.Group("/user")
	controllers.SetupControllerAndRoutes(&usersAPI)


	_ = app.Listen(":8080")
}
