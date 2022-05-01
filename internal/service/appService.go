package service

import (
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/google/uuid"
)

type AppService interface {
	// GetAll returns all the applications registered in its container.
	GetAll() []*model.Application

	// Get returns the application corresponding to the given `id`, or set
	// error to `container.ErrAppNotFound` if the id doen't exists.
	Get(id uuid.UUID) (*model.Application, error)

	// MarshalJSON should return the list of all registered apps in JSON.
	MarshalJSON() ([]byte, error)
}
