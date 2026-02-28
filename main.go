package main

import (
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	protocolID = "/piercing/0.0.0"
)

type Config struct {
	Rings   map[string]string
	Targets []TargetConfig
}

type TargetConfig struct {
	Id    string
	Rings []string
	Path  string
}

type Target struct {
	Config     TargetConfig
	State      TargetState
	LastChange time.Time
}

type TargetState int

type TargetMessage struct {
	TargetId   string
	State      TargetState
	LastChange time.Time
}

const (
	Absent TargetState = iota
	Deleted
	Present
)

func main() {
	var conf Config

	confContent, err := os.ReadFile("./config.toml")
	if err != nil {
		log.Fatal(err)
	}

	_, err = toml.Decode(string(confContent), &conf)
	if err != nil {
		log.Fatal(err)
	}

	p2pSetup(conf)

	select {}
}
