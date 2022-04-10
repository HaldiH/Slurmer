package containers

import (
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/google/uuid"
)

type AppsContainer interface {
	GetAllApp() []*model.Application
	GetApp(id uuid.UUID) (*model.Application, error)
	AddApp(id uuid.UUID, app *model.Application)
	DeleteApp(id uuid.UUID)
}
