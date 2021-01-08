package controllers

import (
	. "../models"
	"time"
)

type UserController struct {

}

func (c *UserController) Get() User {
	return User{
		Username:        "hello",
		Password:        "world",
		PasswordExpired: false,
		PasswordsSent:   0,
		TimeBan:         time.Now(),
		UserDatabase:    UserDatabase{},
	}
}
