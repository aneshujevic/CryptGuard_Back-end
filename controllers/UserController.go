package controllers

import (
	"CryptGuard_Back-end/models"
	"time"
)

type UserController struct {

}

func (c *UserController) Get() models.User {
	return models.User{
		Username:        "hello",
		Password:        "world",
		PasswordExpired: false,
		PasswordsSent:   0,
		TimeBan:         time.Now(),
		UserDatabase:    models.UserDatabase{},
	}
}
