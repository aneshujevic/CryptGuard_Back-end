package models

import (
	"time"
)

type User struct {
	Username string 			`json:"username,validate:required"`
	Password string 			`json:"password,validate:required"`
	PasswordExpired bool 		`json:"password_expired"`
	PasswordsSent int  			`json:"-"`
	TimeBan time.Time 			`json:"-"`
	UserDatabase UserDatabase 	`json:"database"`
}
