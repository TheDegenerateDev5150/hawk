package handlers

import (
	"fmt"
	"{{.ImportPath -}} /svc"
	"os"
	"os/signal"
	"syscall"
)

func SetConfig(cfg svc.Config) svc.Config {
	return cfg
}

func InterruptHandler(errChan chan<- error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	terminateError := fmt.Errorf("%s", <-c)

	// Place whatever shutdown handling you want here

	errChan <- terminateError
}
