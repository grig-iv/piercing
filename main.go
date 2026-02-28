package main

import (
	"log"
	"time"

	"github.com/anthdm/hollywood/actor"
)

type Target struct {
	Config     TargetConfig
	State      TargetState
	LastChange time.Time
}

type TargetMessage struct {
	TargetId   string
	State      TargetState
	LastChange time.Time
}

type TargetState int

const (
	Absent TargetState = iota
	Deleted
	Present
)

func main() {
	conf := loadConfig()

	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}

	fsWatcherPid := engine.Spawn(newFsWatcher, "fsWatcher")

	for _, t := range conf.Targets {
		targetPid := engine.Spawn(targetActorProducer(t, fsWatcherPid), t.Id)
	}

	select {}
}
