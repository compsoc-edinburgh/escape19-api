package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/compsoc-edinburgh/escape19-api/pkg/api"
	"github.com/compsoc-edinburgh/escape19-api/pkg/config"
	"github.com/sirupsen/logrus"

	"github.com/koding/multiconfig"
)

func main() {
	var err error

	m := multiconfig.NewWithPath(os.Getenv("config"))
	cfg := &config.Config{}
	m.MustLoad(cfg)

	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		panic(err)
	}

	logger := logrus.StandardLogger()
	logger.Level = logLevel

	logger.WithFields(logrus.Fields{
		"module": "init",
	}).Info("Starting up infball-api")

	api := api.NewAPI(
		cfg,
		logger,
	)

	go func() {
		logger.WithFields(logrus.Fields{
			"module": "init",
			"bind":   cfg.BindAddress,
		}).Info("Starting the API server")

		if err := api.Start(); err != nil {
			logger.WithFields(logrus.Fields{
				"module": "init",
				"error":  err.Error(),
			}).Fatal("API server failed")
		}
	}()

	// Create a new signal receiver
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Watch for a signal
	<-sc

	// ugly thing to stop ^C from killing alignment
	logger.Out.Write([]byte("\r\n"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := api.Shutdown(ctx); err != nil {
		logger.WithFields(logrus.Fields{
			"module": "init",
			"error":  err.Error(),
		}).Fatal("Failed to close the API server")
	}

	logger.WithFields(logrus.Fields{
		"module": "init",
	}).Info("infball-api has shut down.")
}
