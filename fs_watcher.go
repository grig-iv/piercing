package main

import (
	"log"

	"github.com/anthdm/hollywood/actor"
	"github.com/fsnotify/fsnotify"
)

type fsWatcher struct {
	watcher *fsnotify.Watcher
}

type watchPath struct{ Path string }

func newFsWatcher() actor.Receiver {
	return &fsWatcher{}
}

func (a *fsWatcher) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		w, err := fsnotify.NewWatcher()
		if err != nil {
			panic(err)
		}
		a.watcher = w

	case actor.Started:
		go a.watchLoop(ctx.Engine(), ctx.PID())
		log.Println("FS Watcher started")

	case watchPath:
		if err := a.watcher.Add(msg.Path); err != nil {
			log.Printf("Error adding path %s: %v", msg.Path, err)
		}

	case *fsnotify.Event:
		log.Printf("File Event: %s (Op: %s)", msg.Name, msg.Op)

	case actor.Stopped:
		a.watcher.Close()
		log.Println("FS Watcher stopped")
	}
}

func (a *fsWatcher) watchLoop(engine *actor.Engine, pid *actor.PID) {
	for {
		select {
		case event, ok := <-a.watcher.Events:
			if !ok {
				return
			}
			engine.Send(pid, &event)

		case err, ok := <-a.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("FS Error: %v", err)
		}
	}
}
