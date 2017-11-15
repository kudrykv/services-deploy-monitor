package logging

import (
	"github.com/Sirupsen/logrus"
	"os"
)

var logger *logrus.Logger
var entry *logrus.Entry

func init() {
	logger = logrus.New()
	logger.Out = os.Stdout
	logger.Formatter = &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FieldMap: logrus.FieldMap{
			"msg":   "message",
			"time":  "eventTime",
			"level": "severity",
		},
	}

	entry = logger.WithFields(logrus.Fields{
		"service": "gith-sc",
	})
}

func Level() logrus.Level {
	return logger.Level
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return entry.WithFields(fields)
}

//noinspection GoUnusedExportedFunction
func Debug(args ...interface{}) {
	entry.Debug(args...)
}

//noinspection GoUnusedExportedFunction
func Info(args ...interface{}) {
	entry.Info(args...)
}

//noinspection GoUnusedExportedFunction
func Error(args ...interface{}) {
	entry.Error(args...)
}
