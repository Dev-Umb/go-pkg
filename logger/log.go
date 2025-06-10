// Package logger **/
/**
* @QQ: 1149558764
* @Email: i@umb.ink
 */

package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go.elastic.co/apm"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TraceIDKey 用于在上下文中存储 trace id 的键
const TraceIDKey = "traceID"

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

// generateTraceID 生成traceID
func generateTraceID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GetTraceID 从上下文中获取traceID
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// SetTraceID 将traceID设置到上下文中
func SetTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// getOrGenerateTraceID 获取或生成traceID
func getOrGenerateTraceID(ctx context.Context) (context.Context, string) {
	// 从context中获取traceID
	traceID := GetTraceID(ctx)

	// 如果没有traceID，则生成一个新的
	if traceID == "" {
		traceID = generateTraceID()
		// 将新生成的traceID设置到context中
		ctx = SetTraceID(ctx, traceID)
	}

	return ctx, traceID
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
	ctx, traceID := getOrGenerateTraceID(ctx)
	logger.With(zap.String("trace_id", traceID)).Sugar().Debug(args...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	logger.With(zap.String("trace_id", traceID)).Sugar().Debugf(format, args...)
}

func Info(ctx context.Context, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	logger.With(zap.String("trace_id", traceID)).Sugar().Info(args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	logger.With(zap.String("trace_id", traceID)).Sugar().Infof(format, args...)
}

func Warn(ctx context.Context, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	logger.With(zap.String("trace_id", traceID)).Sugar().Warn(args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	logger.With(zap.String("trace_id", traceID)).Sugar().Warnf(format, args...)
}

func Error(ctx context.Context, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			fmt.Printf("%+v", err)
		}
	}
	logger.With(zap.String("trace_id", traceID)).Sugar().Error(args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	// 打印err的堆栈信息
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			fmt.Printf("%+v", err)
		}
	}
	logger.With(zap.String("trace_id", traceID)).Sugar().Errorf(format, args...)
}

func Panic(ctx context.Context, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	logger.With(zap.String("trace_id", traceID)).Sugar().Panic(args...)
}

func Panicf(ctx context.Context, format string, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	logger.With(zap.String("trace_id", traceID)).Sugar().Panicf(format, args...)
}

func Fatal(ctx context.Context, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	logger.With(zap.String("trace_id", traceID)).Sugar().Fatal(args...)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	ctx, traceID := getOrGenerateTraceID(ctx)
	// 打印堆栈
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			fmt.Printf("%+v", err)
		}
	}
	logger.With(zap.String("trace_id", traceID)).Sugar().Fatalf(format, args...)
}

// 兼容性方法 - 没有context参数的版本，这些方法会创建一个新的context
func DebugWithoutCtx(args ...interface{}) {
	Debug(context.Background(), args...)
}

func DebugfWithoutCtx(format string, args ...interface{}) {
	Debugf(context.Background(), format, args...)
}

func InfoWithoutCtx(args ...interface{}) {
	Info(context.Background(), args...)
}

func InfofWithoutCtx(format string, args ...interface{}) {
	Infof(context.Background(), format, args...)
}

func WarnWithoutCtx(args ...interface{}) {
	Warn(context.Background(), args...)
}

func WarnfWithoutCtx(format string, args ...interface{}) {
	Warnf(context.Background(), format, args...)
}

func ErrorWithoutCtx(args ...interface{}) {
	Error(context.Background(), args...)
}

func ErrorfWithoutCtx(format string, args ...interface{}) {
	Errorf(context.Background(), format, args...)
}

func PanicWithoutCtx(args ...interface{}) {
	Panic(context.Background(), args...)
}

func PanicfWithoutCtx(format string, args ...interface{}) {
	Panicf(context.Background(), format, args...)
}

func FatalWithoutCtx(args ...interface{}) {
	Fatal(context.Background(), args...)
}

func FatalfWithoutCtx(format string, args ...interface{}) {
	Fatalf(context.Background(), format, args...)
}
