package controllers

import (
	"CryptGuard_Back-end/database"
	"CryptGuard_Back-end/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"time"
)

type UserController struct {
	collection *mongo.Collection
}

var userControllerInstance *UserController

func GetUserControllerInstance() *UserController {
	if userControllerInstance == nil {
		userControllerInstance = &UserController{}
		if userControllerInstance.collection == nil {
			client := (database.GetInstance()).Client
			if client == nil {
				panic("Could not get database client.")
			}

			userControllerInstance.collection = client.Database(database.Name).Collection("users")
			if userControllerInstance.collection == nil {
				panic("Could not get users collection")
			}
		}
	}
	return userControllerInstance
}

func SetupControllerAndRoutes(userRoute *fiber.Router) {
	uc := GetUserControllerInstance()

	(*userRoute).Get("/profile", uc.GetUser)
	(*userRoute).Get("/database", uc.GetPasswordDatabase)
	(*userRoute).Post("/database", uc.PostPasswordDatabase)
}

// public accessible handlers
func (uc *UserController) RegisterUser(ctx *fiber.Ctx) error {
	reqUsername := ctx.FormValue("username")
	reqEmail := ctx.FormValue("email")

	if reqEmail == "" || reqUsername == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Username or email missing.",
		})
	}

	var err error
	var dbUser models.User

	err = uc.collection.FindOne(
		context.TODO(),
		bson.M{
			"$or":      []bson.M{{"email": reqEmail}, {"username": reqUsername}},
		}).Decode(&dbUser)

	if err == nil {
		response, _ := json.Marshal("message: Username/email already taken.")
		return ctx.Status(fiber.StatusBadRequest).Send(response)
	}

	user := models.User{
		Username:         reqUsername,
		Email:            reqEmail,
		Password:         "",
		PasswordExpired:  true,
		LoginAttempts:    0,
		TimeBan:          time.Time{},
		PasswordDatabase: models.PasswordDatabaseModel{},
	}

	_, err = uc.collection.InsertOne(context.TODO(), user)
	if err != nil {
		response, _ := json.Marshal("message: Internal server error.")
		return ctx.Status(fiber.StatusInternalServerError).Send(response)
	}

	response, _ := json.Marshal("message: Successfully registered.")
	err = ctx.Send(response)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return err
}

func (uc *UserController) RequestLoginUser(ctx *fiber.Ctx) error {
	username := ctx.FormValue("username")
	password, _ := GenerateRandomString(12)

	var user models.User
	err := uc.collection.FindOne(
		context.TODO(),
		bson.M{"Username": username}).Decode(user)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User does not exist."})
	}

	if !time.Now().After(user.TimeBan) {
		return ctx.JSON(fiber.Map{"message": "You have been temporarily banned."})
	}

	err = uc.collection.FindOneAndUpdate(
		context.TODO(),
		user,
		bson.M{"Password": password, "PasswordExpired": false}).Decode(user)

	if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// simulates sending to email
	log.Print("User: " + username + "\nPassword: " + password)

	return ctx.JSON(fiber.Map{
		"message": "Password sent to email successfully!",
	})
}

func (uc *UserController) LoginUser(ctx *fiber.Ctx) error {
	username := ctx.FormValue("username")
	password := ctx.FormValue("password")

	var foundUser models.User

	err := uc.collection.FindOne(
		context.TODO(),
		bson.M{"Username": username, "Password": password, "PasswordExpired": false},
	).Decode(&foundUser)

	if !time.Now().After(foundUser.TimeBan) {
		return ctx.JSON(fiber.Map{"message": "You have been temporarily banned."})
	}

	if foundUser.PasswordExpired == true {
		return ctx.JSON(fiber.Map{"message": "Password has expired, you should request a new one."})
	}

	if err != nil {
		var timeBan time.Time
		_ = uc.collection.FindOne(
			context.TODO(),
			bson.M{"Username": username}).Decode(&foundUser)

		if foundUser.LoginAttempts >= 5 {
			timeBan = time.Now().Add(time.Minute * 30)
		}

		_ = uc.collection.FindOneAndUpdate(
			context.TODO(),
			foundUser,
			bson.M{"$inc": bson.M{"LoginAttempts": 1}, "TimeBan": timeBan})

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Bad credentials"})
	}

	// successful login
	_ = uc.collection.FindOneAndUpdate(
		context.TODO(),
		foundUser,
		bson.M{"PasswordExpired": true, "LoginAttempts": 0}).Decode(&foundUser)

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = foundUser.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	t, err := token.SignedString([]byte("supersecretkey"))
	if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.JSON(fiber.Map{"token": t})
}

// logged in only handlers
func (uc *UserController) GetUser(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	expiryDate := time.Unix(claims["exp"].(int64), 0)
	var foundUser models.User

	if err := uc.collection.FindOne(context.TODO(), bson.M{"Username": username}).Decode(&foundUser); err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.JSON(fiber.Map{
		"username":             foundUser.Username,
		"password_expiry_date": expiryDate,
		"password_expired":     foundUser.PasswordExpired,
	})
}

func (uc *UserController) GetPasswordDatabase(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	var foundUser models.User

	err := uc.collection.FindOne(
		context.TODO(),
		bson.M{"Username": username}).Decode(&foundUser)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error finding user with username " + username + ".",
		})
	}

	err = ctx.SendFile(fmt.Sprintf("../user_databases/%s", foundUser.PasswordDatabase.Filename), true)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "No database found.",
		})
	}

	return err
}

func (uc *UserController) PostPasswordDatabase(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "No file uploaded.",
		})
	}

	filename := uuid.New()
	err = ctx.SaveFile(file, fmt.Sprintf("../user_databases/%s", filename))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error while saving file.",
		})
	}

	var foundUser models.User
	err = uc.collection.FindOneAndUpdate(
		context.TODO(),
		bson.M{"Username": username},
		bson.M{"$set": bson.M{
			"PasswordDatabase.$.filename":  filename,
			"PasswordDatabase.$.timestamp": time.Now().Unix(),
		}}).Decode(&foundUser)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database file update error.",
		})
	}

	err = os.Remove(fmt.Sprintf("../user_databases/%s", foundUser.PasswordDatabase.Filename))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database file replace error.",
		})
	}

	return err
}
