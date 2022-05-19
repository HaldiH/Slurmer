package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/internal/rest"
	log "github.com/sirupsen/logrus"
)

var cfg = &appconfig.Config{
	Slurmer: appconfig.Slurmer{
		IP:           "0.0.0.0",
		Port:         "8080",
		TemplatesDir: "templates",
		Connector:    "slurmcli",
	},
}

func init() {
	cfgFile := flag.String("c", "config.yml", "Location of the slurmer config file")
	flag.Parse()

	if err := appconfig.FillConfYaml(*cfgFile, cfg); err != nil {
		panic(err)
	}

	if err := cfg.SaveConfYaml(*cfgFile); err != nil {
		panic(err)
	}

	if err := appconfig.FillConfEnv(cfg); err != nil {
		panic(err)
	}

	var formatter log.Formatter
	switch cfg.Slurmer.Logs.Format {
	case "text":
		formatter = &log.TextFormatter{}
	case "json":
		formatter = &log.JSONFormatter{}
	default:
		formatter = &log.TextFormatter{}
	}

	var output io.Writer
	if cfg.Slurmer.Logs.Stdout || cfg.Slurmer.Logs.Output == "" {
		output = os.Stdout
	} else {
		if err := os.MkdirAll(
			filepath.Dir(cfg.Slurmer.Logs.Output),
			os.ModeDir); err != nil {
			panic(err)
		}
		if f, err := os.OpenFile(cfg.Slurmer.Logs.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0440); err != nil {
			panic(err)
		} else {
			output = f
		}
	}

	var level log.Level
	switch cfg.Slurmer.Logs.Level {
	case "trace":
		level = log.TraceLevel
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warning":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	case "fatal":
		level = log.FatalLevel
	case "panic":
		level = log.PanicLevel
	default:
		level = log.InfoLevel
	}

	log.SetFormatter(formatter)
	log.SetOutput(output)
	log.SetLevel(level)
}

func main() {
	log.Debug(*cfg)
	server, err := rest.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	err = server.Listen()
	if err != nil {
		log.Fatal(err)
	}
}
