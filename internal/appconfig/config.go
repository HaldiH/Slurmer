package appconfig

import (
	"errors"
	"os"
	"reflect"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Slurmrest struct {
		URL string `yaml:"url"`
	} `yaml:"slurmrest"`
	Slurmer struct {
		IP           string         `yaml:"ip"`
		Port         string         `yaml:"port"`
		WorkingDir   string         `yaml:"working_dir"`
		TemplatesDir string         `yaml:"templates_dir"`
		Connector    string         `yaml:"connector"`
		Applications []*Application `yaml:"applications"`
		Logs         struct {
			Format string `yaml:"format"`
			Stdout bool   `yaml:"stdout"`
			Output string `yaml:"output"`
			Level  string `yaml:"level"`
		} `yaml:"logs"`
	} `yaml:"slurmer"`
	OIDC struct {
		Issuer   string `yaml:"issuer"`
		Audience string `yaml:"audience"`
	}
}

type Application struct {
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
	UUID  string `yaml:"uuid"`
}

func FillConfYaml(filename string, config *Config) error {
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

func FillConfEnv(config *Config) error {
	confMap := map[any]string{
		&config.Slurmrest.URL: "SLURMREST_URL",

		&config.Slurmer.IP:           "SLURMER_IP",
		&config.Slurmer.Port:         "SLURMER_PORT",
		&config.Slurmer.WorkingDir:   "SLURMER_WD",
		&config.Slurmer.TemplatesDir: "SLURMER_TEMPLATES_DIR",
		&config.Slurmer.Connector:    "SLURMER_CONNECTOR",
		&config.Slurmer.Logs.Format:  "SLURMER_LOGS_FORMAT",
		&config.Slurmer.Logs.Level:   "SLURMER_LOGS_LEVEL",
		&config.Slurmer.Logs.Stdout:  "SLURMER_LOGS_STDOUT",
		&config.Slurmer.Logs.Output:  "SLURMER_LOGS_OUTPUT",
	}

	for ptr, key := range confMap {
		if val, b := os.LookupEnv(key); b {
			ptrRv := reflect.ValueOf(ptr)
			if ptrRv.Kind() != reflect.Pointer {
				return errors.New("Invalid value type: " + ptrRv.Type().Name())
			}

			rv := ptrRv.Elem()
			switch rv.Kind() {
			case reflect.String:
				rv.SetString(val)
			case reflect.Bool:
				rv.SetBool(val != "" && val != "no" && val != "false")
			default:
				return errors.New("Not implemented config type: " + rv.Type().Name())
			}
		}
	}
	return nil
}

func (c *Config) SaveConfYaml(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return err
	}
	return nil
}

func (c *Config) AddApplication(name string) {
	c.Slurmer.Applications = append(c.Slurmer.Applications, &Application{
		Name: name,
		UUID: uuid.NewString(),
	})
}
