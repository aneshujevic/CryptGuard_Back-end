package main

import (
	"CryptGuard_Back-end/controllers"
	"github.com/gofiber/fiber/v2"

	jwtware "github.com/gofiber/jwt/v2"
)

func main() {
	app := fiber.New()

	usersAPI := app.Group("/api")
	usersAPI.Post("/register", controllers.UserControllerInstance.RegisterUser)
	usersAPI.Post("/request-login", controllers.UserControllerInstance.RequestLoginUser)
	usersAPI.Post("/login", controllers.UserControllerInstance.LoginUser)
	usersAPI.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte("supersecretkey"),
	}))
	controllers.SetupControllerAndRoutes(&usersAPI)


	_ = app.Listen(":8080")
}
