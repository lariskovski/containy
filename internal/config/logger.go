package config

import (
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func init() {
	Log.SetFormatter(&logrus.TextFormatter{
		// TimestampFormat: "2006-01-02 15:04:05",
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	Log.SetLevel(logrus.InfoLevel)
}
