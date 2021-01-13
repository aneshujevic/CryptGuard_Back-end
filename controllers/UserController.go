package controllers

import (
	"CryptGuard_Back-end/database"
	"CryptGuard_Back-end/models"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type UserController struct {
	collection *mongo.Collection
}

var userController *UserController

func SetupControllerAndRoutes(userRoute *fiber.Router) {
	userController = &UserController{}
	client := (database.GetInstance()).Client
	if client == nil {
		panic("Could not get database client.")
	}

	userController.collection = client.Database(database.Name).Collection("users")
	if userController.collection == nil {
		panic("Could not get users collection")
	}

	(*userRoute).Get("/profile", userController.GetUser)
	(*userRoute).Post("/register", userController.PostUser)
	(*userRoute).Get("/database", userController.GetPasswordDatabase)
	(*userRoute).Post("/database", userController.PostPasswordDatabase)
}

func (uc *UserController) GetUser(ctx *fiber.Ctx) error {
	return ctx.JSON( models.User {
		Username:          "hello",
		Password:          "world",
		PasswordExpired:   false,
		PasswordsSent:     0,
		TimeBan:           time.Now(),
		PasswordsDatabase: models.PasswordDatabaseModel{},
	})
}

func (uc *UserController) PostUser(ctx *fiber.Ctx) error {
	user := models.User{
		Username:          ctx.Get("username"),
		Password:          "",
		PasswordExpired:   true,
		PasswordsSent:     0,
		TimeBan:           time.Time{},
		PasswordsDatabase: models.PasswordDatabaseModel{},
	}

	_, err := uc.collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return err
}

func (uc *UserController) GetPasswordDatabase(ctx *fiber.Ctx) error {
	return nil
}

func (uc *UserController) PostPasswordDatabase(ctx *fiber.Ctx)  error {
	file, err := ctx.FormFile("file")
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = ctx.SaveFile(file, fmt.Sprintf("./user_databases/%s", file.Filename))
	if err != nil {
		log.Fatal(err)
		return err
	}

	return err
}

func (uc *UserController) PutPasswordDatabase(ctx *fiber.Ctx) error {
	return nil
}
