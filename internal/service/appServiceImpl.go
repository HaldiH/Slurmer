package service

import (
	"encoding/json"
	"fmt"

	"github.com/ShinoYasx/Slurmer/internal/containers"
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/google/uuid"
)

type appServiceImpl struct {
	apps containers.AppsContainer
}

func NewAppServiceImpl(apps containers.AppsContainer) AppService {
	return &appServiceImpl{apps: apps}
}

func (s *appServiceImpl) GetAll() []*model.Application {
	return s.apps.GetAllApp()
}

func (s *appServiceImpl) Get(id uuid.UUID) (*model.Application, error) {
	return s.apps.GetApp(id)
}

func (s *appServiceImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.apps)
}

func (s *appServiceImpl) String() string {
	return fmt.Sprint(s.apps)
}
