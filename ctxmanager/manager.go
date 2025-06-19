// Package ctxmanager 提供统一的context管理功能
package ctxmanager

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// TraceIDKey 用于在上下文中存储 trace id 的键
const TraceIDKey = "traceID"

// ContextManager 上下文管理器
type ContextManager struct {
	mu sync.RWMutex
}

var (
	instance *ContextManager
	once     sync.Once
)

// getInstance 获取单例实例
func getInstance() *ContextManager {
	once.Do(func() {
		instance = &ContextManager{}
	})
	return instance
}

// generateTraceID 生成traceID
func generateTraceID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// NewContext 创建一个新的带有traceID的context
func NewContext() context.Context {
	return NewContextWithParent(context.Background())
}

// NewContextWithParent 基于父context创建一个新的带有traceID的context
func NewContextWithParent(parent context.Context) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	// 检查父context是否已经有traceID
	if traceID := GetTraceID(parent); traceID != "" {
		return parent
	}

	// 生成新的traceID并设置到context中
	traceID := generateTraceID()
	return context.WithValue(parent, TraceIDKey, traceID)
}

// NewContextWithTraceID 创建一个带有指定traceID的context
func NewContextWithTraceID(traceID string) context.Context {
	return NewContextWithTraceIDAndParent(context.Background(), traceID)
}

// NewContextWithTraceIDAndParent 基于父context创建一个带有指定traceID的context
func NewContextWithTraceIDAndParent(parent context.Context, traceID string) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, TraceIDKey, traceID)
}

// GetTraceID 从context中获取traceID
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// SetTraceID 将traceID设置到context中，返回新的context
func SetTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// EnsureTraceID 确保context中有traceID，如果没有则生成一个
func EnsureTraceID(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if GetTraceID(ctx) == "" {
		traceID := generateTraceID()
		return context.WithValue(ctx, TraceIDKey, traceID)
	}

	return ctx
}

// WithTimeout 创建一个带有超时和traceID的context
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	// 确保有traceID
	parent = EnsureTraceID(parent)

	return context.WithTimeout(parent, timeout)
}

// WithCancel 创建一个带有取消功能和traceID的context
func WithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	// 确保有traceID
	parent = EnsureTraceID(parent)

	return context.WithCancel(parent)
}

// WithDeadline 创建一个带有截止时间和traceID的context
func WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	// 确保有traceID
	parent = EnsureTraceID(parent)

	return context.WithDeadline(parent, deadline)
}
