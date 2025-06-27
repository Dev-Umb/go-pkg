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
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/volcengine/volc-sdk-golang/service/tls"
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

	// 火山引擎TLS配置
	TLSConfig *TLSConfig
}

// TLSConfig 火山引擎TLS配置
type TLSConfig struct {
	Enabled         bool   // 是否启用TLS
	Endpoint        string // TLS服务端点
	AccessKeyID     string // 访问密钥ID
	AccessKeySecret string // 访问密钥Secret
	Token           string // 临时访问令牌，可选
	Region          string // 区域
	TopicID         string // 日志主题ID
	Source          string // 日志来源标识
}

// TLSWriter 火山引擎TLS日志写入器
type TLSWriter struct {
	client   tls.Client
	config   *TLSConfig
	logChan  chan []byte
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
	closed   bool
}

// NewTLSWriter 创建新的TLS写入器
func NewTLSWriter(config *TLSConfig) (*TLSWriter, error) {
	if config == nil || !config.Enabled {
		return nil, errors.New("TLS config is nil or disabled")
	}

	// 创建TLS客户端
	client := tls.NewClient(config.Endpoint, config.AccessKeyID, config.AccessKeySecret, config.Token, config.Region)

	writer := &TLSWriter{
		client:   client,
		config:   config,
		logChan:  make(chan []byte, 1000), // 缓冲1000条日志
		stopChan: make(chan struct{}),
	}

	// 启动后台goroutine处理日志发送
	writer.wg.Add(1)
	go writer.processLogs()
	//writer.processLogs()

	return writer, nil
}

// Write 实现io.Writer接口
func (w *TLSWriter) Write(p []byte) (n int, err error) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return 0, errors.New("TLS writer is closed")
	}
	w.mu.RUnlock()

	// 非阻塞写入，如果缓冲区满了就丢弃
	select {
	case w.logChan <- append([]byte(nil), p...):
		return len(p), nil
	default:
		// 缓冲区满了，直接返回成功避免阻塞主程序
		return len(p), nil
	}
}

// Sync 实现zapcore.WriteSyncer接口
func (w *TLSWriter) Sync() error {
	return nil
}

// Close 关闭写入器
func (w *TLSWriter) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	w.mu.Unlock()

	close(w.stopChan)
	w.wg.Wait()
	close(w.logChan)
	return nil
}

// processLogs 处理日志发送的后台goroutine
func (w *TLSWriter) processLogs() {
	defer w.wg.Done()

	ticker := time.NewTicker(1 * time.Second) // 每秒批量发送一次
	defer ticker.Stop()

	var logBuffer [][]byte
	const maxBatchSize = 100 // 最大批量大小

	for {
		select {
		case <-w.stopChan:
			// 发送剩余的日志
			if len(logBuffer) > 0 {
				w.sendLogs(logBuffer)
			}
			return

		case logData := <-w.logChan:
			logBuffer = append(logBuffer, logData)
			// 如果达到批量大小，立即发送
			if len(logBuffer) >= maxBatchSize {
				w.sendLogs(logBuffer)
				logBuffer = logBuffer[:0]
			}

		case <-ticker.C:
			// 定时发送
			if len(logBuffer) > 0 {
				w.sendLogs(logBuffer)
				logBuffer = logBuffer[:0]
			}
		}
	}
}

// sendLogs 发送日志到TLS
func (w *TLSWriter) sendLogs(logBuffer [][]byte) {
	if len(logBuffer) == 0 {
		return
	}

	logs := make([]tls.Log, 0, len(logBuffer))
	for _, logData := range logBuffer {
		// 解析JSON日志内容
		logContent := []tls.LogContent{
			{
				Key:   "message",
				Value: string(logData),
			},
			{
				Key:   "timestamp",
				Value: fmt.Sprintf("%d", time.Now().UnixMilli()),
			},
		}

		// 添加服务名称
		if serviceName := os.Getenv("ELASTIC_APM_SERVICE_NAME"); serviceName != "" {
			logContent = append(logContent, tls.LogContent{
				Key:   "service_name",
				Value: serviceName,
			})
		}

		logs = append(logs, tls.Log{
			Contents: logContent,
		})
	}

	// 发送日志
	_, err := w.client.PutLogsV2(&tls.PutLogsV2Request{
		TopicID:      w.config.TopicID,
		CompressType: "lz4",
		Source:       w.config.Source,
		FileName:     "go-logger",
		Logs:         logs,
	})

	if err != nil {
		// 记录错误，但不阻塞程序运行
		fmt.Printf("Failed to send logs to TLS: %v\n", err)
	}
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
	var syncWriters []zapcore.WriteSyncer

	// 文件写入器
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
	syncWriters = append(syncWriters, zapcore.AddSync(lumberJackLogger))

	// TLS写入器
	if config.TLSConfig != nil && config.TLSConfig.Enabled {
		tlsWriter, err := NewTLSWriter(config.TLSConfig)
		if err != nil {
			fmt.Printf("Failed to create TLS writer: %v\n", err)
		} else {
			syncWriters = append(syncWriters, zapcore.AddSync(tlsWriter))
		}
	}

	return zapcore.NewMultiWriteSyncer(syncWriters...)
}

func getLogFile(config *Config) string {
	fileFormat := time.Now().Format(config.FileFormat)
	fileName := strings.Join([]string{
		config.FilePrefix,
		fileFormat,
		"log"}, ".")
	return path.Join(config.FilePath, fileName)
}
