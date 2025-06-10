# 变更日志

## [v1.1.0] - 2024-12-XX

### 新增功能

#### 🎛️ 订阅管理功能

- **订阅ID管理**: 所有订阅方法现在返回唯一的订阅ID，便于后续管理
- **精确取消订阅**: 支持通过订阅ID精确取消特定订阅
- **批量取消订阅**: 支持按频道取消所有相关订阅
- **全量取消订阅**: 支持一键取消所有活跃订阅
- **订阅状态查询**: 可以查看当前所有活跃订阅的详细信息

#### 新增API方法

**Client接口新增方法:**
- `Unsubscribe(subscriptionId string) error` - 取消指定订阅
- `UnsubscribeByChannel(channel string) error` - 取消指定频道的所有订阅
- `UnsubscribeAll() error` - 取消所有订阅
- `GetActiveSubscriptions() []SubscriptionInfo` - 获取活跃订阅信息

**全局方法新增:**
- `notification.Unsubscribe(subscriptionId string) error`
- `notification.UnsubscribeByChannel(channel string) error`
- `notification.UnsubscribeAll() error`
- `notification.GetActiveSubscriptions() []SubscriptionInfo`

#### 新增数据结构

```go
// SubscriptionInfo 订阅信息
type SubscriptionInfo struct {
    ID      string `json:"id"`       // 订阅唯一ID
    Channel string `json:"channel"`  // 订阅的频道
    OpenId  string `json:"open_id"`  // 用户OpenId（如果适用）
    Active  bool   `json:"active"`   // 是否活跃
}
```

### 破坏性变更

⚠️ **API签名变更**: 所有订阅方法的返回值从 `error` 变更为 `(string, error)`，第一个返回值为订阅ID。

**受影响的方法:**
- `SubscribeKickOff` - 现在返回 `(string, error)`
- `SubscribeLogin` - 现在返回 `(string, error)`
- `SubscribeLogout` - 现在返回 `(string, error)`
- `SubscribeCustomEvent` - 现在返回 `(string, error)`
- `SubscribeKickOffTyped` - 现在返回 `(string, error)`
- `SubscribeLoginTyped` - 现在返回 `(string, error)`
- `SubscribeLogoutTyped` - 现在返回 `(string, error)`
- `SubscribeMultipleKickOff` - 现在返回 `([]string, error)`

### 改进

- **资源管理**: 改进了订阅的生命周期管理，防止资源泄漏
- **线程安全**: 所有订阅管理操作都是线程安全的
- **错误处理**: 增强了取消订阅时的错误处理和日志记录
- **自动清理**: 客户端关闭时自动取消所有活跃订阅

### 示例

#### 基本用法

```go
// 订阅事件并获取订阅ID
subId, err := client.SubscribeKickOff("user123", func(payload string) {
    log.Printf("收到事件: %s", payload)
})

// 取消订阅
err = client.Unsubscribe(subId)

// 查看活跃订阅
subscriptions := client.GetActiveSubscriptions()
```

#### 批量管理

```go
// 取消用户的所有踢下线订阅
err := client.UnsubscribeByChannel("user:kickoff:user123")

// 取消所有订阅
err := client.UnsubscribeAll()
```

### 迁移指南

如果您正在从旧版本升级，需要更新订阅方法的调用方式：

**旧版本:**
```go
err := client.SubscribeKickOff("user123", handler)
```

**新版本:**
```go
subId, err := client.SubscribeKickOff("user123", handler)
// 保存 subId 以便后续取消订阅
```

### 新增示例

- `examples/unsubscribe_usage/main.go` - 完整的取消订阅功能演示

---

## [v1.0.0] - 2024-06-XX

### 初始版本

- 基于Redis Pub/Sub的用户通知系统
- 支持踢下线、登录、退出事件
- 全局单例模式和实例模式
- 类型安全的事件处理器
- 自定义日志器支持 