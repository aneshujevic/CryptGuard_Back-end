package main

import (
	"CryptGuard_Back-end/controllers"
	"CryptGuard_Back-end/database"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	usersAPI := app.Group("/user")
	controllers.SetupControllerAndRoutes(&usersAPI)
	database.GetInstance()
	database.DestroyInstance()

	_ = app.Listen(":8080")
}
