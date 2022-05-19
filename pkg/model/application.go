package model

import "github.com/google/uuid"

type Application struct {
	Name        string    `json:"name"`
	AccessToken string    `json:"access_token"`
	Directory   string    `json:"directory"`
	Id          uuid.UUID `json:"id"`
}
