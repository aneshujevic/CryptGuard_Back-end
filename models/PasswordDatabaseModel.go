package models

import (
	"mime/multipart"
	"time"
)

type PasswordDatabaseModel struct {
	Filename string 	`json:"name,validate:required"`
	File multipart.File `json:"file,validate:required"`
	Timestamp time.Time `json:"-"`
}
