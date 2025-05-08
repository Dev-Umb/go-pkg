// Package logger **/
/**
* @Author: chenhao29
* @Date: 2025/3/3 13:58
* @QQ: 1149558764
* @Email: i@umb.ink
 */

package logger

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ApmConfig struct {
	FilePath    string
	MaxFileSize int
	MaxBackups  int
	MaxAge      int
	Compress    bool
	LogLevel    string
	FileFormat  string
	FilePrefix  string
}

func (log *kLogger) Fatal(args ...interface{}) {
	s := fmt.Sprint(args...)
	e := apm.CaptureError(log.ctx, errors.New(s))
	e.Send()

	log.logger.Fatal(s)
}

func (log *kLogger) Fatalf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	e := apm.CaptureError(log.ctx, errors.New(s))
	e.Send()

	log.logger.Fatal(s)
}

func (log *kLogger) Panic(args ...interface{}) {
	s := fmt.Sprint(args...)
	e := apm.CaptureError(log.ctx, errors.New(s))
	e.Send()

	log.logger.Panic(s)
}

func (log *kLogger) Panicf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	e := apm.CaptureError(log.ctx, errors.New(s))
	e.Send()

	log.logger.Panic(s)
}

func (log *kLogger) Error(args ...interface{}) {
	s := fmt.Sprint(args...)
	e := apm.CaptureError(log.ctx, errors.New(s))
	e.Send()

	log.logger.Error(s)
}

func (log *kLogger) Errorf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	e := apm.CaptureError(log.ctx, errors.New(s))
	e.Send()

	log.logger.Error(s)
}

func (log *kLogger) Info(args ...interface{}) {
	s := fmt.Sprint(args...)
	log.logger.Info(s)
}

func (log *kLogger) Infof(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	log.logger.Info(s)
}

func (log *kLogger) Debug(args ...interface{}) {
	s := fmt.Sprint(args...)
	log.logger.Debug(s)
}

func (log *kLogger) Debugf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	log.logger.Debug(s)
}

func (log *kLogger) With(fields ...interface{}) *kLogger {
	log.logger = log.logger.Sugar().With(fields...).Desugar()
	return log
}

func (log *kLogger) WithTrace() *kLogger {
	tx := apm.TransactionFromContext(log.ctx)
	traceContext := tx.TraceContext()
	return log.With(zap.String("trace.id", traceContext.Trace.String()), zap.String("transaction.id", traceContext.Span.String()))
}

func WithContext(ctx context.Context) *kLogger {
	l := logger.With(apmzap.TraceContext(ctx)...)
	return &kLogger{logger: l, ctx: ctx}
}

func (log *kLogger) WithLabel(labels map[string]string) *kLogger {
	span := apm.SpanFromContext(log.ctx)
	for k, v := range labels {
		span.Context.SetLabel(k, v)
		log.logger = log.logger.With(zap.String(k, v))
	}
	return log
}

func getJsonEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getEncoderConfig() zapcore.EncoderConfig {
	zapConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + caller.TrimmedPath() + "]")
		},
	}
	return zapConfig
}

func getLogWriter(config *Config) zapcore.WriteSyncer {
	if config.FilePath == "" {
		config.FilePath = "./log"
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 100
	}
	if config.MaxAge == 0 {
		config.MaxAge = 30
	}

	lumberJackLogger := &lumberjack.Logger{
		Filename:   getLogFile(config),
		MaxSize:    config.MaxFileSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func getLogFile(config *Config) string {
	fileFormat := time.Now().Format(config.FileFormat)
	fileName := strings.Join([]string{
		config.FilePrefix,
		fileFormat,
		"log"}, ".")
	return path.Join(config.FilePath, fileName)
}
