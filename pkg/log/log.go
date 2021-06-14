// https://github.com/onrik/gorm-logrus
package log

import (
	"context"
	"errors"
	"time"

	filename "github.com/keepeye/logrus-filename"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type Logger struct {
	Logrus                *log.Logger
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
}

func (l *Logger) SetLogLevel(logLevel string) {
	switch logLevel {
	case "debug":
		l.Logrus.SetLevel(log.DebugLevel)
	case "info":
	default:
		l.Logrus.SetLevel(log.InfoLevel)
	}
}

func New() *Logger {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	filenameHook := filename.NewHook()
	filenameHook.Field = "line"
	logger.AddHook(filenameHook)
	return &Logger{
		Logrus:                logger,
		SkipErrRecordNotFound: true,
	}
}

func (l *Logger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *Logger) Info(ctx context.Context, s string, args ...interface{}) {
	l.Logrus.WithContext(ctx).Infof(s, args)
}

func (l *Logger) Warn(ctx context.Context, s string, args ...interface{}) {
	l.Logrus.WithContext(ctx).Warnf(s, args)
}

func (l *Logger) Error(ctx context.Context, s string, args ...interface{}) {
	l.Logrus.WithContext(ctx).Errorf(s, args)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := log.Fields{}
	if l.SourceField != "" {
		fields[l.SourceField] = utils.FileWithLineNum()
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		fields[log.ErrorKey] = err
		l.Logrus.WithContext(ctx).WithFields(fields).Errorf("%s [%s]", sql, elapsed)
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		l.Logrus.WithContext(ctx).WithFields(fields).Warnf("%s [%s]", sql, elapsed)
		return
	}

	l.Logrus.WithContext(ctx).WithFields(fields).Debugf("%s [%s]", sql, elapsed)
}
