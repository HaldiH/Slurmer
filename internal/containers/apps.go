package containers

import (
	"errors"

	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/google/uuid"
)

var ErrAppNotFound error = errors.New("cannot find an app with such uuid")
var ErrAppAlreadyExists error = errors.New("app already registered")

type AppsContainer interface {
	GetAllApp() ([]*model.Application, error)
	GetApp(id uuid.UUID) (*model.Application, error)
	AddApp(id uuid.UUID, app *model.Application) error
	DeleteApp(id uuid.UUID) error
}
