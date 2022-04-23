package appconfig

import (
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Slurmrest struct {
		URL string `yaml:"url"`
	} `yaml:"slurmrest"`
	Slurmer struct {
		IP           string         `yaml:"ip"`
		Port         uint16         `yaml:"port"`
		WorkingDir   string         `yaml:"working_dir"`
		Connector    string         `yaml:"connector"`
		Applications []*Application `yaml:"applications"`
		Logs         struct {
			Format string `yaml:"format"`
			Stdout bool   `yaml:"stdout"`
			Output string `yaml:"output"`
			Level  string `yaml:"level"`
		} `yaml:"logs"`
	} `yaml:"slurmer"`
}

type Application struct {
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
	UUID  string `yaml:"uuid"`
}

func MakeYamlConf(filename string, config *Config) error {
	f, err := os.Open(filename)
	if err != nil {
		return nil // We don't want to throw an error il the file doesn't exist.
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(config); err != nil {
		return err
	}

	for _, app := range config.Slurmer.Applications {
		if app.UUID == "" {
			app.UUID = uuid.NewString()
		}
	}

	return nil
}

func SaveYamlConf(filename string, config *Config) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	if err := encoder.Encode(config); err != nil {
		return err
	}
	return nil
}

func AddApplication(name string, config *Config) {
	config.Slurmer.Applications = append(config.Slurmer.Applications, &Application{
		Name: name,
		UUID: uuid.NewString(),
	})
}
