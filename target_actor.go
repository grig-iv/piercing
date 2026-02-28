package main

import (
	"log"

	"github.com/anthdm/hollywood/actor"
	"github.com/fsnotify/fsnotify"
)

type targetActor struct {
	config       TargetConfig
	fsWatcherPid *actor.PID
}

func targetActorProducer(config TargetConfig, fsWatcherPid *actor.PID) actor.Producer {
	return func() actor.Receiver {
		return &targetActor{
			config:       config,
			fsWatcherPid: fsWatcherPid,
		}
	}
}

func (a *targetActor) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		ctx.Send(a.fsWatcherPid, watchPath{a.config.Path})
		log.Printf("Target actor of id '%s' started", a.config.Id)

	case *fsnotify.Event:
		log.Printf("File Event: %s (Op: %s)", msg.Name, msg.Op)

	case actor.Stopped:
		log.Printf("Target actor of id '%s' stopped", a.config.Id)
	}
}
