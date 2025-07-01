# Context Manager

统一的Go Context管理工具包，自动携带traceID，完美适配logger库。

## 特性

- ✅ 单例模式，无需管理实例
- ✅ 自动生成和管理traceID
- ✅ 完全兼容标准context包
- ✅ 支持所有标准context操作（WithTimeout、WithCancel、WithDeadline）
- ✅ 与现有logger库无缝集成

## 快速开始

### 基本用法

```go
import "github.com/Dev-Umb/go-pkg/ctxmanager"

// 创建一个新的带traceID的context
ctx := ctxmanager.NewContext()

// 获取traceID
traceID := ctxmanager.GetTraceID(ctx)
fmt.Printf("TraceID: %s\n", traceID)
```

### 与Logger集成

```go
import (
    "github.com/Dev-Umb/go-pkg/ctxmanager"
    "github.com/Dev-Umb/go-pkg/logger"
)

func handleRequest() {
    // 创建带traceID的context
    ctx := ctxmanager.NewContext()
    
    // 直接使用，logger会自动获取traceID
    logger.Info(ctx, "处理请求开始")
    logger.Errorf(ctx, "发生错误: %v", err)
}
```

## API 文档

### 核心方法

#### `NewContext() context.Context`
创建一个新的带有traceID的context。

```go
ctx := ctxmanager.NewContext()
```

#### `NewContextWithParent(parent context.Context) context.Context`
基于父context创建新context，如果父context已有traceID则复用，否则生成新的。

```go
ctx := ctxmanager.NewContextWithParent(parentCtx)
```

#### `NewContextWithTraceID(traceID string) context.Context`
创建带有指定traceID的context。

```go
ctx := ctxmanager.NewContextWithTraceID("custom-trace-123")
```

#### `EnsureTraceID(ctx context.Context) context.Context`
确保context中有traceID，如果没有则自动生成。

```go
ctx = ctxmanager.EnsureTraceID(ctx)
```

### 工具方法

#### `GetTraceID(ctx context.Context) string`
从context中获取traceID。

```go
traceID := ctxmanager.GetTraceID(ctx)
```

#### `SetTraceID(ctx context.Context, traceID string) context.Context`
设置traceID到context中，返回新的context。

```go
ctx = ctxmanager.SetTraceID(ctx, "new-trace-id")
```

### 标准Context操作

所有方法都会自动确保context中包含traceID：

#### `WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc)`
```go
ctx, cancel := ctxmanager.WithTimeout(parentCtx, 5*time.Second)
defer cancel()
```

#### `WithCancel(parent context.Context) (context.Context, context.CancelFunc)`
```go
ctx, cancel := ctxmanager.WithCancel(parentCtx)
defer cancel()
```

#### `WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc)`
```go
deadline := time.Now().Add(10 * time.Second)
ctx, cancel := ctxmanager.WithDeadline(parentCtx, deadline)
defer cancel()
```

## 使用场景

### 1. HTTP请求处理

```go
func handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
    // 创建带traceID的context
    ctx := ctxmanager.NewContextWithParent(r.Context())
    
    // 后续所有操作都使用这个context
    logger.Info(ctx, "开始处理HTTP请求")
    
    // 调用业务逻辑
    result, err := businessLogic(ctx)
    if err != nil {
        logger.Error(ctx, "业务逻辑执行失败", err)
        return
    }
    
    logger.Info(ctx, "HTTP请求处理完成")
}
```

### 2. 数据库操作

```go
func queryDatabase(ctx context.Context) error {
    // 确保context有traceID
    ctx = ctxmanager.EnsureTraceID(ctx)
    
    logger.Info(ctx, "开始数据库查询")
    
    // 执行数据库操作...
    
    logger.Info(ctx, "数据库查询完成")
    return nil
}
```

### 3. 微服务调用

```go
func callMicroservice(ctx context.Context) error {
    // 确保有traceID用于链路追踪
    ctx = ctxmanager.EnsureTraceID(ctx)
    
    traceID := ctxmanager.GetTraceID(ctx)
    
    // 将traceID添加到HTTP头部
    req.Header.Set("X-Trace-ID", traceID)
    
    logger.Infof(ctx, "调用微服务，traceID: %s", traceID)
    
    // 执行HTTP调用...
    
    return nil
}
```

## 注意事项

1. **Context不可变性**：所有操作都会返回新的context，原context不会被修改
2. **线程安全**：所有方法都是线程安全的
3. **性能**：traceID生成使用crypto/rand，保证唯一性
4. **兼容性**：完全兼容标准context包，可以无缝替换

## 与现有代码集成

如果你已经有使用context的代码，只需要简单替换：

```go
// 原来的代码
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

// 替换为
ctx, cancel := ctxmanager.WithTimeout(context.Background(), 5*time.Second)
```

这样就能自动获得traceID功能，无需修改其他代码。 