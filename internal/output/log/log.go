package log

import (
	"github.com/sirupsen/logrus"
	"io"
)

type Option func(l *logrus.Logger)

func WithLevel(level logrus.Level) Option {
	return func(logger *logrus.Logger) {
		logger.SetLevel(level)
	}
}

func WithOutput(output io.Writer) Option {
	return func(logger *logrus.Logger) {
		logger.SetOutput(output)
	}
}

func NewJSONLogger(opts ...Option) *logrus.Logger {
	logger := logrus.New()
	for _, f := range opts {
		f(logger)
	}

	logger.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp:  false,
		DisableHTMLEscape: false,
		PrettyPrint:       false,
	})

	return logger
}
