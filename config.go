package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Rings   map[string]string
	Targets []TargetConfig
}

type TargetConfig struct {
	Id    string
	Type  TargetType
	Rings []string
	Path  string
}

type TargetType string

const (
	File TargetType = "file"
	Dir  TargetType = "dir"
)

func loadConfig() Config {
	var conf Config

	confContent, err := os.ReadFile("./config.toml")
	if err != nil {
		log.Fatal(err)
	}

	_, err = toml.Decode(string(confContent), &conf)
	if err != nil {
		log.Fatal(err)
	}

	for i, target := range conf.Targets {
		conf.Targets[i] = TargetConfig{
			Id:    target.Id,
			Type:  target.Type,
			Rings: target.Rings,
			Path:  normolizePath(target.Path),
		}
	}

	return conf
}

func normolizePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		return filepath.Join(home, path[2:])
	} else {
		path, err := filepath.Abs(path)
		if err != nil {
			log.Fatal(err)
		}

		return path
	}
}
