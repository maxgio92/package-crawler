package log

import (
	"github.com/sirupsen/logrus"
	"os"
)

func NewLogger(level logrus.Level) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp:  false,
		DisableHTMLEscape: false,
		PrettyPrint:       false,
	})
	logger.SetLevel(level)

	return logger
}
