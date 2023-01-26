package main

import (
	"akt-docker/container"
	"os"

	log "github.com/sirupsen/logrus"
)

func Run(tty bool, command string) {
	// clone 一个 namespace
	parent := container.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	parent.Wait()
	os.Exit(-1)
}
