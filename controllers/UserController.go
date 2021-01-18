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
	"strings"
	"time"
)

type UserController struct {
	collection *mongo.Collection
}

const CURR_PATH = "/home/goofy/playground/Go/CryptGuard_Back-end"

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
	reqEmail := strings.TrimSpace(ctx.FormValue("email"))
	reqUsername := strings.TrimSpace(ctx.FormValue("username"))

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
			"$or": []bson.M{{"email": reqEmail}, {"username": reqUsername}},
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
		bson.M{"username": username}).Decode(&user)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User does not exist."})
	}

	if !time.Now().After(user.TimeBan) {
		return ctx.JSON(fiber.Map{"message": "You have been temporarily banned."})
	}

	_, err = uc.collection.UpdateOne(
		context.TODO(),
		user,
		bson.M{"$set": bson.M{"password": password, "passwordexpired": false}})

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
	username := strings.TrimSpace(ctx.FormValue("username"))
	password := strings.TrimSpace(ctx.FormValue("password"))

	var foundUser models.User

	err := uc.collection.FindOne(
		context.TODO(),
		bson.M{"username": username},
	).Decode(&foundUser)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "No user with that username.",
		})
	}

	if !time.Now().After(foundUser.TimeBan) {
		return ctx.JSON(fiber.Map{"message": "You have been temporarily banned."})
	}

	if foundUser.PasswordExpired == true {
		return ctx.JSON(fiber.Map{"message": "Password has expired, you should request a new one."})
	}

	err = uc.collection.FindOne(
		context.TODO(),
		bson.M{"username": username, "password": password},
	).Decode(&foundUser)
	if err != nil {
		var timeBan time.Time
		_ = uc.collection.FindOne(
			context.TODO(),
			bson.M{"username": username}).Decode(&foundUser)

		if foundUser.LoginAttempts >= 5 {
			fmt.Println()
			timeBan = time.Now().Add(time.Minute * 30)
		}

		_, err = uc.collection.UpdateOne(
			context.TODO(),
			bson.M{"username": foundUser.Username},
			bson.M{
				"$set": bson.M{"timeban": timeBan},
			},
		)
		if err != nil {
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}

		_, err = uc.collection.UpdateOne(
			context.TODO(),
			bson.M{"username": foundUser.Username},
			bson.M{
				"$inc": bson.M{"loginattempts": 1},
			})
		if err != nil {
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Bad credentials"})
	}

	// successful login
	_, err = uc.collection.UpdateOne(
		context.TODO(),
		bson.M{"username": username},
		bson.M{"$set": bson.M{"passwordexpired": true, "loginattempts": 0}})

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

// log in protected handlers
func (uc *UserController) GetUser(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	expiryDate := time.Unix(int64(claims["exp"].(float64)), 0)
	var foundUser models.User

	if err := uc.collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&foundUser); err != nil {
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
		bson.M{"username": username}).Decode(&foundUser)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error finding user with username " + username + ".",
		})
	}

	err = ctx.SendFile(fmt.Sprintf(CURR_PATH+"/user_databases/%s", foundUser.PasswordDatabase.Filename), true)
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
	err = ctx.SaveFile(file, fmt.Sprintf(CURR_PATH+"/user_databases/%s", filename))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error while saving file.",
		})
	}

	var foundUser models.User
	err = uc.collection.FindOneAndUpdate(
		context.TODO(),
		bson.M{"username": username},
		bson.M{"$set": bson.M{
			"passworddatabase.filename":  filename.String(),
			"passworddatabase.timestamp": time.Now().Unix(),
		}}).Decode(&foundUser)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database file update error.",
		})
	}

	if foundUser.PasswordDatabase.Filename != "" {
		err = os.Remove(fmt.Sprintf(CURR_PATH+"/user_databases/%s", foundUser.PasswordDatabase.Filename))
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database file replace error.",
			})
		}
	}

	return err
}
