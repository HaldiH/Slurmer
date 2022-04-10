package model

import "github.com/google/uuid"

type Application struct {
	AccessToken string
	Directory   string
	Id          uuid.UUID
}
