package containers

import (
	"errors"

	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/google/uuid"
)

var ErrAppNotFound error = errors.New("Cannot find an app with such uuid")

type AppsContainer interface {
	GetAllApp() []*model.Application
	GetApp(id uuid.UUID) (*model.Application, error)
	AddApp(id uuid.UUID, app *model.Application)
	DeleteApp(id uuid.UUID)
}
