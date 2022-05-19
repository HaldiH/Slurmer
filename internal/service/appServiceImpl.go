package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/internal/containers"
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/ShinoYasx/Slurmer/pkg/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func InitAppDir(app *model.Application, tplDir string) error {
	app.Directory = filepath.Join(appconfig.AppsDir, app.Id.String())
	jobsDir := filepath.Join(app.Directory, "jobs")
	templatesDir := path.Join(app.Directory, "templates")

	// Will create app and jobs directory under /applications/{uuid}/jobs/
	if err := os.MkdirAll(jobsDir, os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(templatesDir, os.ModePerm); err != nil {
		return err
	}

	if _, err := os.Stat(tplDir); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err := utils.CopyDirectory(tplDir, templatesDir, false); err != nil {
		log.Error("Error in templates directory copy")
		return err
	}

	return nil
}

type appServiceImpl struct {
	apps       containers.AppsContainer
	configPath string
	tplDir     string
}

func NewAppService(apps containers.AppsContainer, configPath string, tplDir string) (AppService, error) {
	return &appServiceImpl{apps: apps, configPath: configPath, tplDir: tplDir}, nil
}

func (s *appServiceImpl) GetAll() ([]*model.Application, error) {
	return s.apps.GetAllApp()
}

func (s *appServiceImpl) Get(id uuid.UUID) (*model.Application, error) {
	return s.apps.GetApp(id)
}

func (s *appServiceImpl) Add(app *model.Application) error {
	accessToken, err := appconfig.GenAppToken(rand.Reader)
	if err != nil {
		return err
	}

	app.Id = uuid.New()
	app.AccessToken = accessToken
	if err := InitAppDir(app, s.tplDir); err != nil {
		return err
	}

	newApp := appconfig.Application{
		Name:  app.Name,
		Token: app.AccessToken,
		UUID:  app.Id.String(),
	}

	var cfg appconfig.Config

	read, err := os.Open(s.configPath)
	if err != nil {
		return err
	}

	decoder := yaml.NewDecoder(read)
	if err := decoder.Decode(&cfg); err != nil {
		return err
	}
	read.Close()

	cfg.Slurmer.Applications = append(cfg.Slurmer.Applications, &newApp)

	write, err := os.Create(s.configPath)
	if err != nil {
		return err
	}
	encoder := yaml.NewEncoder(write)
	if err := encoder.Encode(&cfg); err != nil {
		return err
	}
	write.Close()

	if err := s.apps.AddApp(app.Id, app); err != nil {
		return err
	}

	return nil
}

func (s *appServiceImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.apps)
}

func (s *appServiceImpl) String() string {
	return fmt.Sprint(s.apps)
}
