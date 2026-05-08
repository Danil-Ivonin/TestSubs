package logger

import (
	"fmt"
	"strings"

	"github.com/Danil-Ivonin/TestSubs/internal/config"
	"github.com/sirupsen/logrus"
)

func New(cfg config.LogConfig) (*logrus.Logger, error) {
	log := logrus.New()

	level, err := logrus.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		return nil, fmt.Errorf("parsing log level: %w", err)
	}
	log.SetLevel(level)

	switch strings.ToLower(cfg.Format) {
	case "", "text":
		log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	default:
		return nil, fmt.Errorf("unsupported log format: %s", cfg.Format)
	}

	return log, nil
}
