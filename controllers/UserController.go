package controllers

import (
	"CryptGuard_Back-end/models"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"
)

func SetupUserRoutes(userRoute *fiber.Router) {
	(*userRoute).Get("", Get)
	(*userRoute).Post("", Post)
	(*userRoute).Put("", Put)
}

func Put(ctx *fiber.Ctx) error {
	// TODO: Implement logic for updating
	file, err := ctx.FormFile("file")
	if err != nil {
		return err
	}

	return ctx.SaveFile(file, fmt.Sprintf("./user_databases/%s", file.Filename))
}

func Post(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return err
	}

	return ctx.SaveFile(file, fmt.Sprintf("./user_databases/%s", file.Filename))
}

func Get(ctx *fiber.Ctx) error {
	return ctx.JSON( models.User {
			Username:        "hello",
			Password:        "world",
			PasswordExpired: false,
			PasswordsSent:   0,
			TimeBan:         time.Now(),
			UserDatabase:    models.UserDatabase{},
		})
}
