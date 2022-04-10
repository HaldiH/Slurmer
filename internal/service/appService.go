package service

import (
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/google/uuid"
)

type AppService interface {
	GetAll() []*model.Application
	Get(id uuid.UUID) (*model.Application, error)

	// MarshallJSON should return the list of all registered apps in JSON.
	MarshalJSON() ([]byte, error)
}
