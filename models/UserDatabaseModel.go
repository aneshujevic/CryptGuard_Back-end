package models

import (
	"mime/multipart"
	"time"
)

type UserDatabase struct {
	Filename string 	`json:"name"`
	File multipart.File `json:"file"`
	Timestamp time.Time `json:"-"`
}
