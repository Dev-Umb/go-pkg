# Nacos客户端使用指南

本文档介绍如何使用单例模式的Nacos客户端获取配置和服务发现。

## 架构说明

Nacos组件抽离为两个主要部分：
1. 配置管理 - 使用单例模式实现，全局共享一个配置客户端实例
2. 服务注册发现 - 使用单例模式实现，全局共享一个服务发现客户端实例

## 配置客户端API

### 获取配置

从Nacos获取配置值：

```go
import "game-room-manager-server/internal/pkg/nacos_sdk"

// 获取配置，使用默认分组
value, err := nacos.GetConfigValue("配置ID", nacos.GetDefaultGroup())
if err != nil {
    // 处理错误
}

// 使用指定分组
value, err := nacos.GetConfigValue("配置ID", "自定义分组")
if err != nil {
    // 处理错误
}
```

### 监听配置变更

监听配置变更并处理：

```go
import "game-room-manager-server/internal/pkg/nacos_sdk"

// 定义配置变更处理函数
handler := func(newValue string) {
    // 处理新的配置值
    config.SomeValue = newValue
    // 可能需要重启服务或更新内部状态
}

// 开始监听配置变更，使用默认分组
err := nacos.ListenConfigChange("配置ID", nacos.GetDefaultGroup(), handler)
if err != nil {
    // 处理错误
}

// 使用指定分组
err := nacos.ListenConfigChange("配置ID", "自定义分组", handler)
if err != nil {
    // 处理错误
}
```

## 服务发现客户端API

### 获取健康服务实例

获取一个健康的服务实例：

```go
import "game-room-manager-server/internal/pkg/nacos_sdk"

// 获取健康的服务实例
instance, err := nacos.GetHealthyInstance("服务名称", "分组名称")
if err != nil {
    // 处理错误
}

// 使用实例信息
fmt.Printf("服务实例: %s:%d\n", instance.Ip, instance.Port)
```

### 获取所有服务实例

获取指定服务的所有实例：

```go
import "game-room-manager-server/internal/pkg/nacos_sdk"

// 获取所有健康的服务实例
instances, err := nacos.GetAllInstances("服务名称", "分组名称", true)
if err != nil {
    // 处理错误
}

// 获取所有服务实例，包括不健康的
allInstances, err := nacos.GetAllInstances("服务名称", "分组名称", false)
if err != nil {
    // 处理错误
}
```

### 创建gRPC客户端

使用服务发现创建gRPC客户端：

```go
import "game-room-manager-server/internal/pkg/nacos_sdk"

// 创建gRPC客户端的工厂函数
func newServiceClient(conn *grpc.ClientConn) pb.ServiceClient {
    return pb.NewServiceClient(conn)
}

// 获取gRPC客户端
client, err := nacos.GetGRPCClient("服务名称", "分组名称", newServiceClient)
if err != nil {
    // 处理错误
}

// 使用客户端调用远程方法
response, err := client.RemoteMethod(ctx, request)
```

### 订阅服务变更

监听服务实例变更：

```go
import "game-room-manager-server/internal/pkg/nacos_sdk"

// 定义服务变更回调函数
callback := func(instances []model.Instance, err error) {
    if err != nil {
        // 处理错误
        return
    }
    
    // 处理服务实例变更
    fmt.Printf("服务实例数量变更为: %d\n", len(instances))
    for i, instance := range instances {
        fmt.Printf("实例 %d: %s:%d, 健康状态: %v\n", 
            i+1, instance.Ip, instance.Port, instance.Healthy)
    }
}

// 订阅服务变更
err := nacos.SubscribeService("服务名称", "分组名称", callback)
if err != nil {
    // 处理错误
}
```

### 注册服务实例

注册自己的服务实例：

```go
import "game-room-manager-server/internal/pkg/nacos_sdk"

// 注册服务实例
metadata := map[string]string{"version": "1.0.0"}
success, err := nacos.RegisterServiceInstance(
    "服务名称", 
    "127.0.0.1", 
    8080, 
    "分组名称",
    metadata,
)
if err != nil {
    // 处理错误
}

// 服务关闭时注销实例
defer func() {
    nacos.DeregisterServiceInstance("服务名称", "127.0.0.1", 8080, "分组名称")
}()
```

## 注意事项

1. 客户端是全局单例，无需手动初始化，首次调用API时会自动初始化
2. 配置变更回调函数会在后台goroutine中执行，注意并发安全
3. 在监听配置变更时，如果配置变更需要重启服务，建议使用restart.RestartService()确保优雅重启
4. 服务发现客户端的连接可能会暂时断开，但SDK会自动重连
5. 服务注册时使用临时实例，确保服务宕机时能自动从Nacos移除 

## 配置使用示例

以下是一个完整的示例，展示如何获取和监听Redis配置：

```go
package main

import (
    "game-room-manager-server/config"
    "game-room-manager-server/internal/pkg/nacos"
    "game-room-manager-server/pkg/logger"
    "game-room-manager-server/pkg/restart"
)

func initRedisConfig() error {
    // 获取Redis配置
    redisConfig, err := nacos.GetConfigValue("redis.conf", nacos.GetDefaultGroup())
    if err != nil {
        logger.Errorf("获取Redis配置失败: %v", err)
        return err
    }

    // 应用配置
    config.RedisURL = redisConfig
    
    // 监听配置变更
    err = nacos.ListenConfigChange("redis.conf", nacos.GetDefaultGroup(), func(data string) {
        logger.Infof("Redis配置变更: %s", data)
        config.RedisURL = data
        // 如果需要重启服务以应用新配置
        restart.RestartService()
    })
    if err != nil {
        logger.Warnf("监听Redis配置失败: %v", err)
    }
    
    return nil
} 