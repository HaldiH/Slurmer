package rest

import (
	"github.com/ShinoYasx/Slurmer/internal/containers"
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/ShinoYasx/Slurmer/pkg/utils"
	"github.com/google/uuid"
)

type appsMap map[uuid.UUID]*model.Application

func NewAppsMap() containers.AppsContainer {
	c := make(appsMap)
	return &c
}

func (m *appsMap) GetAllApp() ([]*model.Application, error) {
	apps := []*model.Application{}
	for _, v := range *m {
		apps = append(apps, v)
	}
	return apps, nil
}

func (m *appsMap) GetApp(id uuid.UUID) (*model.Application, error) {
	app := (*m)[id]
	if app == nil {
		return nil, containers.ErrAppNotFound
	}
	return app, nil
}

func (m *appsMap) AddApp(id uuid.UUID, app *model.Application) error {
	if _, exists := (*m)[id]; exists {
		return containers.ErrAppAlreadyExists
	}
	(*m)[id] = app
	return nil
}

func (m *appsMap) DeleteApp(id uuid.UUID) error {
	delete(*m, id)
	return nil
}

func (m *appsMap) MarshalJSON() ([]byte, error) { return utils.MapToJSONArray(*m) }
