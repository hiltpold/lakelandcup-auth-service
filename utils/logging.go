package utils

import (
	"go.uber.org/zap"
)

var zapLog *zap.Logger

func init() {
	var err error
	config := zap.NewDevelopmentConfig()
	enccoderConfig := zap.NewDevelopmentEncoderConfig()
	//zapcore.TimeEncoderOfLayout("Jan _2 15:04:05.000000000")
	enccoderConfig.StacktraceKey = "" // to hide stacktrace info
	config.EncoderConfig = enccoderConfig

	zapLog, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

func Info(message string, fields ...zap.Field) {
	zapLog.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	zapLog.Debug(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	zapLog.Error(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	zapLog.Fatal(message, fields...)
}
