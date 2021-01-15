package models

import (
	"time"
)

type PasswordDatabaseModel struct {
	Filename string 	`json:"name,validate:required"`
	Timestamp time.Time `json:"-"`
}
