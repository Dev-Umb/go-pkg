package main

import (
	"context"

	"github.com/Dev-Umb/go-pkg/logger"
)

func main() {
	// 初始化logger
	_, err := logger.Use(&logger.Config{
		ApmConfig: logger.ApmConfig{
			LogLevel: "debug",
		},
	})
	if err != nil {
		panic(err)
	}

	println("=== 测试1: 模拟HTTP请求处理 - 没有预设traceId ===")
	handleRequest(context.Background())

	println("\n=== 测试2: 模拟HTTP请求处理 - 有预设traceId ===")
	ctxWithTraceId := logger.SetTraceID(context.Background(), "http-request-trace-12345")
	handleRequest(ctxWithTraceId)

	println("\n=== 测试3: 兼容性方法测试 ===")
	logger.InfoWithoutCtx("使用兼容性方法的日志")
	logger.ErrorWithoutCtx("使用兼容性方法的错误日志")
}

// 模拟HTTP请求处理函数
func handleRequest(ctx context.Context) {
	logger.Info(ctx, "开始处理HTTP请求")

	// 调用业务逻辑
	processBusinessLogic(ctx)

	// 调用数据库
	queryDatabase(ctx)

	logger.Info(ctx, "HTTP请求处理完成")
}

// 模拟业务逻辑处理
func processBusinessLogic(ctx context.Context) {
	logger.Debug(ctx, "开始执行业务逻辑")
	logger.Infof(ctx, "业务参数验证: %s", "参数正确")

	// 模拟一些错误场景
	if false { // 这里可以改为true来测试错误日志
		logger.Error(ctx, "业务逻辑执行失败")
		return
	}

	logger.Debug(ctx, "业务逻辑执行完成")
}

// 模拟数据库查询
func queryDatabase(ctx context.Context) {
	logger.Debug(ctx, "开始数据库查询")
	logger.Infof(ctx, "执行SQL: %s", "SELECT * FROM users WHERE id = 1")
	logger.Debug(ctx, "数据库查询完成")
}
