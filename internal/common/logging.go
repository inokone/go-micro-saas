package common

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func initLogging(config *LogConfig) {
	if config.PrettyLog {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.SetLevel(log.DebugLevel)
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Failed to parse log level, default is debug.")
	} else {
		log.SetLevel(level)
	}
}

func GinLogFormatter(param gin.LogFormatterParams) string {
	log.WithFields(log.Fields{
		"client_ip":  param.ClientIP,
		"time_stamp": param.TimeStamp.Format(time.RFC3339),
		"method":     param.Method,
		"path":       param.Path,
		"status":     param.StatusCode,
		"latency":    param.Latency,
		"user_agent": param.Request.UserAgent(),
		"error":      param.ErrorMessage,
	}).Info("request log")
	return ""
}

func LoggerMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// After the request is handled
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// Log structured data using Logrus
		logger.WithFields(log.Fields{
			"timestamp":  endTime.Format(time.RFC3339),
			"status":     c.Writer.Status(),
			"latency_ms": latency.Milliseconds(),
			"client_ip":  c.ClientIP(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_agent": c.Request.UserAgent(),
		}).Info("Request completed")
	}
}
