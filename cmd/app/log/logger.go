package log

import (
	"fmt"
	"os"

	"github.com/Doom-z/RepClient/cmd/app/cfg"
	"github.com/Doom-z/RepClient/pkg/logger"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

func InitLogger(conf cfg.Log, verbose bool) {
	newLoggerCount := len(conf.Stdout) + len(conf.File)
	if newLoggerCount != 0 {
		for i := 0; i < newLoggerCount; i++ {
			logger.AddLogger(logrus.New())
		}
	}
	nextLoggerIndex := 1

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
		// enable source code line numbers
		// logger.SetReportCaller(true)
	} else {
		logger.SetLevel(cfg.LevelToLogrusLevel(conf.Level))
	}

	for _, stdout := range conf.Stdout {
		currentLogger := logger.DefaultCombinedLogger.GetLogger(nextLoggerIndex)
		switch stdout.Format {
		case cfg.LogFormatJSON:
			fmt.Println("use json")
			currentLogger.SetFormatter(&logrus.JSONFormatter{})
		case cfg.LogFormatText:
			currentLogger.SetFormatter(&logrus.TextFormatter{})
		}
		switch stdout.Output {
		case cfg.LogOutputStdout:
			currentLogger.SetOutput(os.Stdout)
		case cfg.LogOutputStderr:
			currentLogger.SetOutput(os.Stderr)
		}
		nextLoggerIndex++
	}

	for _, file := range conf.File {
		currentLogger := logger.DefaultCombinedLogger.GetLogger(nextLoggerIndex)

		switch file.Format {
		case cfg.LogFormatJSON:
			currentLogger.SetFormatter(&logrus.JSONFormatter{})
		case cfg.LogFormatText:
			currentLogger.SetFormatter(&logrus.TextFormatter{})
		}
		currentLogger.SetOutput(&lumberjack.Logger{
			Filename: file.Path,
			MaxSize:  file.MaxSize,
			MaxAge:   file.MaxAge,
			// MaxBackups: 3,
			Compress: true,
		})
		nextLoggerIndex++
	}

}
