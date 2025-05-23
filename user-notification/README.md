# User Notification SDK

一个基于 Redis Pub/Sub 的用户账户通知 SDK，支持用户登录、退出、踢下线等事件的发布和订阅。支持**全局单例模式**和**实例模式**两种使用方式。

## 特性

- 🚀 **易于使用**: 简单的 API 设计，快速集成
- 🔒 **个性化频道**: 基于用户 OpenID 的专属频道设计
- 📡 **实时通知**: 基于 Redis Pub/Sub 的实时事件推送
- 🎯 **类型安全**: 强类型的事件结构和处理器
- 🔧 **高度可配置**: 支持自定义 Redis 配置和日志器
- ⚡ **高性能**: 连接池和超时控制
- 🛡️ **错误处理**: 完善的错误处理和重连机制
- 🌍 **全局单例**: 支持全局单例模式，方便在整个应用中使用

## 安装

```bash
go get login-server/pkg/user-notification
```

## 使用方式

SDK 提供两种使用方式：

### 1. 全局单例模式（推荐）

适用于大多数应用场景，一次初始化，全局使用。

#### 初始化全局客户端

```go
import notification "login-server/pkg/user-notification"

// 在应用启动时初始化全局客户端
func main() {
    err := notification.InitGlobalClient(
        "redis://localhost:6379",
        notification.WithPassword("password"),
        notification.WithDB(1),
        notification.WithTimeout(10*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 确保程序退出时关闭客户端
    defer notification.CloseGlobalClient()
}
```

#### 在任何地方使用全局方法

```go
// 发布事件
err := notification.PublishKickOff("user123", "platform001", "在其他设备登录")
err := notification.PublishLogin("user123", "platform001", "192.168.1.100", "UserAgent")
err := notification.PublishLogout("user123", "platform001", 3600)

// 订阅事件
err := notification.SubscribeKickOffTyped("user123", func(event notification.UserKickOffEvent) {
    log.Printf("用户 %s 被踢下线: %s", event.OpenId, event.Reason)
})
```

#### 检查初始化状态

```go
if notification.IsGlobalClientInitialized() {
    log.Println("全局客户端已初始化")
}
```

### 2. 实例模式

适用于需要多个独立客户端的场景。

```go
// 创建独立的客户端实例
client, err := notification.NewClient("redis://localhost:6379")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// 使用实例方法
err = client.PublishKickOff("user123", "platform001", "在其他设备登录")
```

## 在现有项目中的集成

### 在登录服务中使用

在您的用户登录逻辑中，当检测到用户已在线时发布踢下线事件：

```go
func (up *userPlatforms) PlatformAuth(...) {
    // ... 现有登录逻辑 ...
    
    if user.LoginStatus == 1 {
        // 用户在线，将被踢下线，使用全局单例的 notification SDK 发布事件
        if !notification.IsGlobalClientInitialized() {
            logger.Errorf("全局通知客户端未初始化，无法发布踢下线事件")
        } else {
            // 发布踢下线事件
            err = notification.PublishKickOff(mapping.OpenId, platformCode, "用户在其他设备登录")
            if err != nil {
                logger.Errorf("发布用户踢下线事件失败: %+v", err)
                // 不中断流程，继续执行登录逻辑
            } else {
                logger.Infof("成功发布用户踢下线事件: UnionId=%s, OpenId=%s, PlatformCode=%s",
                    user.UnionId, mapping.OpenId, platformCode)
            }
        }
    }
    
    // ... 继续登录逻辑 ...
}
```

### 在下游服务中订阅事件

```go
// 在应用启动时订阅相关事件
func InitEventSubscription() {
    // 订阅特定用户的踢下线事件
    err := notification.SubscribeKickOffTyped("user123", func(event notification.UserKickOffEvent) {
        log.Printf("用户 %s 被踢下线，原因: %s", event.OpenId, event.Reason)
        
        // 处理踢下线逻辑
        // 1. 通知前端用户下线
        // 2. 清理用户相关缓存
        // 3. 记录审计日志
        handleUserKickOff(event.OpenId, event.PlatformCode, event.Reason)
    })
    
    if err != nil {
        log.Printf("订阅踢下线事件失败: %v", err)
    }
}

func handleUserKickOff(openId, platformCode, reason string) {
    // 实现具体的踢下线处理逻辑
    log.Printf("处理用户踢下线: OpenId=%s, Platform=%s, Reason=%s", 
        openId, platformCode, reason)
}
```

## 快速开始

### 1. 创建客户端

```go
package main

import (
    "log"
    notification "login-server/pkg/user-notification"
)

func main() {
    // 简单创建
    client, err := notification.NewClient("redis://localhost:6379")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 或者使用选项
    client, err = notification.NewClient(
        "redis://localhost:6379",
        notification.WithPassword("your-password"),
        notification.WithDB(1),
        notification.WithTimeout(10*time.Second),
    )
}
```

### 2. 发布事件

```go
// 发布踢下线事件
err := client.PublishKickOff("user123", "platform001", "在其他设备登录")

// 发布登录事件
err := client.PublishLogin("user123", "platform001", "192.168.1.100", "UserAgent")

// 发布退出事件
err := client.PublishLogout("user123", "platform001", 3600) // 在线1小时
```

### 3. 订阅事件

```go
// 订阅踢下线事件（原始处理器）
err := client.SubscribeKickOff("user123", func(payload string) {
    log.Printf("收到踢下线事件: %s", payload)
})

// 订阅踢下线事件（类型化处理器）
err := client.SubscribeKickOffTyped("user123", func(event notification.UserKickOffEvent) {
    log.Printf("用户 %s 被踢下线: %s", event.OpenId, event.Reason)
})
```

## 详细用法

### 配置选项

```go
client, err := notification.NewClient(
    "redis://localhost:6379",
    notification.WithPassword("password"),        // Redis 密码
    notification.WithDB(1),                       // 数据库编号
    notification.WithTimeout(10*time.Second),     // 连接超时
    notification.WithPoolSize(20),                // 连接池大小
    notification.WithLogger(myLogger),            // 自定义日志器
)
```

### 自定义日志器

```go
type MyLogger struct{}

func (l *MyLogger) Info(args ...interface{}) { /* 实现 */ }
func (l *MyLogger) Infof(format string, args ...interface{}) { /* 实现 */ }
func (l *MyLogger) Error(args ...interface{}) { /* 实现 */ }
func (l *MyLogger) Errorf(format string, args ...interface{}) { /* 实现 */ }
func (l *MyLogger) Warn(args ...interface{}) { /* 实现 */ }
func (l *MyLogger) Warnf(format string, args ...interface{}) { /* 实现 */ }

// 使用自定义日志器
client, err := notification.NewClient(
    "redis://localhost:6379",
    notification.WithLogger(&MyLogger{}),
)
```

### 事件类型

#### 踢下线事件

```go
type UserKickOffEvent struct {
    OpenId       string    `json:"open_id"`
    EventType    string    `json:"event_type"`
    PlatformCode string    `json:"platform_code"`
    Timestamp    time.Time `json:"timestamp"`
    Message      string    `json:"message"`
    Reason       string    `json:"reason,omitempty"`
}
```

#### 登录事件

```go
type UserLoginEvent struct {
    OpenId       string    `json:"open_id"`
    EventType    string    `json:"event_type"`
    PlatformCode string    `json:"platform_code"`
    Timestamp    time.Time `json:"timestamp"`
    Message      string    `json:"message"`
    ClientIP     string    `json:"client_ip,omitempty"`
    UserAgent    string    `json:"user_agent,omitempty"`
}
```

#### 退出事件

```go
type UserLogoutEvent struct {
    OpenId       string    `json:"open_id"`
    EventType    string    `json:"event_type"`
    PlatformCode string    `json:"platform_code"`
    Timestamp    time.Time `json:"timestamp"`
    Message      string    `json:"message"`
    Duration     int64     `json:"duration,omitempty"` // 在线时长（秒）
}
```

### 批量操作

```go
// 批量订阅多个用户的踢下线事件
userIds := []string{"user1", "user2", "user3"}
err := client.SubscribeMultipleKickOff(userIds, func(payload string) {
    log.Printf("收到消息: %s", payload)
})
```

### 自定义事件

```go
// 发布自定义事件
extra := map[string]interface{}{
    "level": 10,
    "score": 1000,
}
err := client.PublishCustomEvent("user123", "game_level_up", "game", "升级", extra)

// 订阅自定义事件
err := client.SubscribeCustomEvent("user:game_level_up:user123", func(payload string) {
    log.Printf("收到游戏事件: %s", payload)
})
```

## 频道设计

SDK 使用基于用户 OpenID 的个性化频道设计：

- 踢下线频道：`user:kickoff:{openId}`
- 登录频道：`user:login:{openId}`
- 退出频道：`user:logout:{openId}`
- 自定义频道：`user:{eventType}:{openId}`

### 优势

1. **精确推送**: 用户只接收自己的事件
2. **提高安全性**: 防止用户接收其他用户的敏感信息
3. **减少资源消耗**: 无需在业务层过滤无关消息
4. **易于扩展**: 支持用户级别的细粒度控制

## 连接管理

```go
// 检查连接状态
if client.IsConnected() {
    log.Println("客户端已连接")
}

// 测试连接
ctx := context.Background()
if err := client.Ping(ctx); err != nil {
    log.Printf("连接测试失败: %v", err)
}

// 关闭连接
defer client.Close()
```

## 错误处理

SDK 提供了完善的错误处理机制：

- 连接失败时自动重试
- 发布失败时返回详细错误信息
- 订阅断开时自动重连
- 所有错误都通过日志器记录

## 性能考虑

- 使用连接池管理 Redis 连接
- 支持配置连接超时和重试次数
- 异步事件处理，不阻塞主流程
- JSON 序列化优化

## 示例

查看 [examples](examples/) 目录获取完整示例：

- [basic_usage.go](examples/basic_usage.go) - 基本用法示例（实例模式）
- [singleton_usage/main.go](examples/singleton_usage/main.go) - 全局单例模式示例
- [downstream_service/main.go](examples/downstream_service/main.go) - 下游服务订阅事件示例
- [advanced_usage.go](examples/advanced_usage.go) - 高级用法示例

## 常见问题

### Q: 如何处理 Redis 连接失败？

A: SDK 会自动重试连接，你可以通过 `WithTimeout` 和 `MaxRetries` 配置重试参数。

### Q: 是否支持集群模式？

A: 当前版本支持单机和哨兵模式，集群模式支持将在后续版本中添加。

### Q: 如何监控事件处理性能？

A: 可以通过自定义日志器记录事件处理时间，或集成 Prometheus 等监控系统。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！ 