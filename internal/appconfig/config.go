package appconfig

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Slurmrest struct {
		URI string `yaml:"url"`
	} `yaml:"slurmrest"`
	Slurmer struct {
		IP   string `yaml:"ip"`
		Port uint16 `yaml:"port"`
	} `yaml:"slurmer"`
}

func MakeYamlConf(filename string, config *Config) error {
	f, err := os.Open("config.yml")
	if err != nil {
		return nil // We don't want to throw an error il the file doesn't exist.
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(config)
	if err != nil {
		return err
	}
	return nil
}
