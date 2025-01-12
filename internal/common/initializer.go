package common

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// InitApp sets up the app including configuration management and logging
func InitApp(path string) *AppConfig {
	config, err := loadConfig(path)
	if err != nil {
		log.WithError(err).Error("Failed to load application configuration.")
		os.Exit(1)
	}
	initLogging(config.Log)
	return config
}
