package logger

import (
	"log"

	"go.uber.org/zap"
)

func Init() (*zap.Logger, error) {

	logr, err := Use(&Config{
		ApmConfig: ApmConfig{
			FilePath:    "/home/web_server/kid-runner/logs/",
			FilePrefix:  "kid-edu-app",
			FileFormat:  "2006-01-02",
			LogLevel:    "info",
			MaxFileSize: 128,
			MaxAge:      30,
			MaxBackups:  3,
			Compress:    true,
		},
	})
	if err != nil {
		log.Fatal("启动日志失败！")
	}

	return logr, nil
}
