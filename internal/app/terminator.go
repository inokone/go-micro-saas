package app

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func listenOS(cancel func()) {
	var osSignals = make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-osSignals
		log.WithField("signal", sig).Info("The application is shutting down...")
		cancel()
	}()
}
