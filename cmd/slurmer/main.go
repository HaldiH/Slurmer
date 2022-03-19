package main

import (
	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/internal/slurmer"
)

func main() {
	var cfg appconfig.Config
	err := appconfig.MakeYamlConf("config.yml", &cfg)
	if err != nil {
		panic(err)
	}

	server, err := slurmer.New(&cfg)
	if err != nil {
		panic(err)
	}
	err = server.Listen()
	if err != nil {
		panic(err)
	}
}
