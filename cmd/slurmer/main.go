package main

import (
	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/internal/server"
)

func main() {
	var cfg appconfig.Config
	err := appconfig.MakeYamlConf("config.yml", &cfg)
	if err != nil {
		panic(err)
	}

	server := server.Server{
		Config: &cfg,
	}

	err = server.Listen()
	if err != nil {
		panic(err)
	}
}
