// Package logger **/
/**
* @Author: chenhao29
* @Date: 2025/3/3 13:59
* @QQ: 1149558764
* @Email: i@umb.ink
 */

package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go.elastic.co/apm"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	ApmConfig
}

type kLogger struct {
	logger *zap.Logger
	ctx    context.Context
}

var logger *zap.Logger

var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"panic": zapcore.PanicLevel,
	"fatal": zapcore.FatalLevel,
}

func init() {
	initGlobalLogger("debug")
}
func initGlobalLogger(logLevel string) {
	var syncWriters []zapcore.WriteSyncer

	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000000"))
	}

	syncWriters = append(syncWriters, zapcore.AddSync(os.Stdout))

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder),
		zapcore.NewMultiWriteSyncer(syncWriters...),
		zap.NewAtomicLevelAt(GetLoggerLevel(logLevel)),
	)

	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func GetLoggerLevel(l string) zapcore.Level {
	if level, ok := levelMap[l]; ok {
		return level
	}
	return zapcore.InfoLevel
}

func Use(config *Config) (*zap.Logger, error) {
	if config == nil {
		return nil, errors.New("config ptr is nil, Init failed")
	}
	if config.LogLevel == "" {
		config.LogLevel = "debug"
	}
	config.LogLevel = strings.TrimSpace(config.LogLevel)
	config.LogLevel = strings.ToLower(config.LogLevel)
	initGlobalLogger(config.LogLevel)

	writeSyncer := getLogWriter(config)
	encoder := getJsonEncoder()
	fileCore := zapcore.NewCore(encoder, writeSyncer, levelMap[config.LogLevel])

	logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, fileCore)
	}))
	apm.DefaultTracer.SetLogger(logger.Sugar())

	fields := withFields()
	logger = logger.With(fields...)
	zap.ReplaceGlobals(logger)
	return logger, nil
}

func withFields() []zap.Field {
	var fields []zap.Field
	projectName := os.Getenv("ELASTIC_APM_SERVICE_NAME")
	if len(projectName) == 0 {
		projectName = os.Args[0]
		projectName = strings.Trim(projectName, "./")
	}
	if len(projectName) == 0 {
		projectName = "please set a projectName"
	}
	fields = append(fields, zap.String("application.name", projectName))
	if AppVersion() != "" {
		fields = append(fields, zap.String("version", AppVersion()))
	}
	if BuildTime() != "" {
		fields = append(fields, zap.String("build.time", BuildTime()))
	}
	if BuildHost() != "" {
		fields = append(fields, zap.String("build.host", BuildHost()))
	}
	if BuildUser() != "" {
		fields = append(fields, zap.String("build.user", BuildUser()))
	}
	if HostName() != "" {
		fields = append(fields, zap.String("hostname", HostName()))
	}
	if GoVersion() != "" {
		fields = append(fields, zap.String("go.version", GoVersion()))
	}
	return fields
}

func Debug(ctx context.Context, args ...interface{}) {
	logger.Sugar().With(ctx).Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Sugar().Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Sugar().Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Sugar().Infof(format, args...)
}

func Warn(args ...interface{}) {
	logger.Sugar().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Sugar().Warnf(format, args...)
}

func Error(args ...interface{}) {
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			fmt.Printf("%+v", err)
		}
	}
	logger.Sugar().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	// 打印err的堆栈信息
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			fmt.Printf("%+v", err)
		}
	}
	logger.Sugar().Errorf(format, args...)
}

func Panic(args ...interface{}) {
	logger.Sugar().Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	logger.Sugar().Panicf(format, args...)
}

func Fatal(args ...interface{}) {
	logger.Sugar().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	// 打印堆栈
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			fmt.Printf("%+v", err)
		}
	}
	logger.Sugar().Fatalf(format, args...)
}
