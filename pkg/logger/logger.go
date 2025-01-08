package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type AppLoggerConfig struct {
	Level      string
	FormatType LoggerFormatType
	Output     LoggerOutput
	// OutputPath is the path to the log file
	//
	// If Output is LoggerOutputFile, OutputPath is the path to the log file, and
	// if Output is LoggerOutputStdout or LoggerOutputStderr, OutputPath is ignored, can be empty
	OutputPath string
}

type LoggerOutput int

const (
	LoggerOutputStdout LoggerOutput = iota
	LoggerOutputStderr
	LoggerOutputFile
	// Not implemented yet
	LoggerOutputDatabase
)

type LoggerLevel string

const (
	LoggerLevelTrace LoggerLevel = "trace"
	LoggerLevelDebug             = "debug"
	LoggerLevelInfo              = "info"
	LoggerLevelWarn              = "warn"
	LoggerLevelError             = "error"
	LoggerLevelFatal             = "fatal"
	LoggerLevelPanic             = "panic"
)

type LoggerFormatType int

const (
	LoggerFormatConsole LoggerFormatType = iota
	LoggerFormatMessageOnly
	LoggerFormatJSON
)

func SetAppLogger(config *AppLoggerConfig) {
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}

	logrus.SetLevel(level)
	var formatter logrus.Formatter

	switch config.FormatType {
	case LoggerFormatConsole:
		formatter = &logrus.TextFormatter{}
	case LoggerFormatMessageOnly:
		formatter = &logrus.TextFormatter{
			DisableTimestamp: true,
		}
	case LoggerFormatJSON:
		formatter = &logrus.JSONFormatter{}
	default:
		formatter = &logrus.TextFormatter{}
	}

	logrus.SetFormatter(formatter)

	switch config.Output {
	case LoggerOutputStdout:
		logrus.SetOutput(os.Stdout)
	case LoggerOutputFile:
		file, err := os.OpenFile(config.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("Failed to open log file: %v", err))
		}
		logrus.SetOutput(file)
	case LoggerOutputStderr:
	default:
		logrus.SetOutput(os.Stderr)
	}
}

func Trace(args ...interface{}) {
	logrus.Trace(args...)
}

func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func Error(args ...interface{}) {
	logrus.Error(args...)
}

func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func Panic(args ...interface{}) {
	logrus.Panic(args...)
}

func Tracef(format string, args ...interface{}) {
	logrus.Tracef(format, args...)
}

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	logrus.Panicf(format, args...)
}

func init() {
	SetAppLogger(&AppLoggerConfig{
		Level:      string(LoggerLevelInfo),
		FormatType: LoggerFormatConsole,
		Output:     LoggerOutputStderr,
	})
}
