# Logger 库

这是一个支持火山引擎TLS（日志服务）的Go日志库，基于zap构建。

## 功能特性

- 支持多种日志级别（Debug、Info、Warn、Error、Panic、Fatal）
- 自动生成和管理TraceID
- 支持文件日志轮转
- 支持火山引擎TLS日志服务
- 保持所有原有接口不变，无感接入

## 快速开始

### 基本使用（不使用TLS）

```go
package main

import (
    "context"
    "github.com/Dev-Umb/go-pkg/logger"
)

func main() {
    // 使用默认配置
    ctx := context.Background()
    logger.Info(ctx, "Hello, World!")
    
    // 或者使用自定义配置
    config := &logger.Config{
        ApmConfig: logger.ApmConfig{
            LogLevel: "info",
            FilePath: "./logs",
            MaxFileSize: 100,
            MaxAge: 30,
        },
    }
    
    logger.Use(config)
    logger.Info(ctx, "Using custom config")
}
```

### 使用火山引擎TLS日志服务

```go
package main

import (
    "context"
    "os"
    "github.com/Dev-Umb/go-pkg/logger"
)

func main() {
    config := &logger.Config{
        ApmConfig: logger.ApmConfig{
            LogLevel: "info",
            FilePath: "./logs",
            MaxFileSize: 100,
            MaxAge: 30,
            TLSConfig: &logger.TLSConfig{
                Enabled:         true,
                Endpoint:        os.Getenv("VOLCENGINE_ENDPOINT"),        // 火山引擎TLS服务端点
                AccessKeyID:     os.Getenv("VOLCENGINE_ACCESS_KEY_ID"),   // 访问密钥ID
                AccessKeySecret: os.Getenv("VOLCENGINE_ACCESS_KEY_SECRET"), // 访问密钥Secret
                Token:           os.Getenv("VOLCENGINE_TOKEN"),           // 临时访问令牌（可选）
                Region:          os.Getenv("VOLCENGINE_REGION"),          // 区域
                TopicID:         os.Getenv("VOLCENGINE_TOPIC_ID"),        // 日志主题ID
                Source:          "my-service",                            // 日志来源标识
            },
        },
    }
    
    logger.Use(config)
    
    ctx := context.Background()
    logger.Info(ctx, "日志将同时写入文件和火山引擎TLS")
    logger.Error(ctx, "错误日志也会发送到TLS")
}
```

### 环境变量配置

建议通过环境变量配置火山引擎TLS相关参数：

```bash
export VOLCENGINE_ENDPOINT="https://tls-cn-beijing.volces.com"
export VOLCENGINE_ACCESS_KEY_ID="your-access-key-id"
export VOLCENGINE_ACCESS_KEY_SECRET="your-access-key-secret"
export VOLCENGINE_TOKEN=""  # 可选
export VOLCENGINE_REGION="cn-beijing"
export VOLCENGINE_TOPIC_ID="your-topic-id"
```

## API 接口

所有原有的日志接口保持不变：

### 带Context的接口
- `Debug(ctx context.Context, args ...interface{})`
- `Debugf(ctx context.Context, format string, args ...interface{})`
- `Info(ctx context.Context, args ...interface{})`
- `Infof(ctx context.Context, format string, args ...interface{})`
- `Warn(ctx context.Context, args ...interface{})`
- `Warnf(ctx context.Context, format string, args ...interface{})`
- `Error(ctx context.Context, args ...interface{})`
- `Errorf(ctx context.Context, format string, args ...interface{})`
- `Panic(ctx context.Context, args ...interface{})`
- `Panicf(ctx context.Context, format string, args ...interface{})`
- `Fatal(ctx context.Context, args ...interface{})`
- `Fatalf(ctx context.Context, format string, args ...interface{})`

### 兼容性接口（不带Context）
- `DebugWithoutCtx(args ...interface{})`
- `DebugfWithoutCtx(format string, args ...interface{})`
- `InfoWithoutCtx(args ...interface{})`
- `InfofWithoutCtx(format string, args ...interface{})`
- `WarnWithoutCtx(args ...interface{})`
- `WarnfWithoutCtx(format string, args ...interface{})`
- `ErrorWithoutCtx(args ...interface{})`
- `ErrorfWithoutCtx(format string, args ...interface{})`
- `PanicWithoutCtx(args ...interface{})`
- `PanicfWithoutCtx(format string, args ...interface{})`
- `FatalWithoutCtx(args ...interface{})`
- `FatalfWithoutCtx(format string, args ...interface{})`

## TLS配置说明

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| Enabled | bool | 是 | 是否启用TLS日志服务 |
| Endpoint | string | 是 | TLS服务端点 |
| AccessKeyID | string | 是 | 访问密钥ID |
| AccessKeySecret | string | 是 | 访问密钥Secret |
| Token | string | 否 | 临时访问令牌 |
| Region | string | 是 | 区域 |
| TopicID | string | 是 | 日志主题ID |
| Source | string | 是 | 日志来源标识 |

## 特性说明

### TraceID管理
- 自动生成和管理TraceID
- 支持从Context中获取和设置TraceID
- 每条日志都会包含TraceID信息

### 批量发送
- TLS日志采用异步批量发送机制
- 默认每秒发送一次或达到100条日志时立即发送
- 避免阻塞主程序运行

### 错误处理
- TLS发送失败不会影响程序正常运行
- 错误信息会输出到标准错误流
- 日志仍会正常写入文件

### 资源管理
- 程序退出时会自动关闭TLS连接
- 发送剩余的缓冲日志
- 释放相关资源

## 注意事项

1. 确保火山引擎TLS服务的网络连通性
2. 正确配置访问密钥和权限
3. TopicID必须是已存在的日志主题
4. 建议在生产环境中通过环境变量配置敏感信息
5. TLS发送是异步的，程序退出前会尽力发送剩余日志 