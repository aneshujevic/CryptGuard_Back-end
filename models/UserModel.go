package models

import (
	"time"
)

type User struct {
	Username         string                `json:"username,validate:required"`
	Email            string                `json:"email,validate:required"`
	Password         string                `json:"password"`
	PasswordExpired  bool                  `json:"password_expired"`
	LoginAttempts    int                   `json:"-"`
	TimeBan          time.Time             `json:"-"`
	PasswordDatabase PasswordDatabaseModel `json:"database"`
}
