package log

import (
	filename "github.com/keepeye/logrus-filename"
	log "github.com/sirupsen/logrus"
)

func New() *log.Logger {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	filenameHook := filename.NewHook()
	filenameHook.Field = "line"
	logger.AddHook(filenameHook)
	return logger
}
