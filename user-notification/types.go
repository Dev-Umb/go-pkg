package notification

import (
	"context"
	"time"
)

// UserEvent 用户事件通用结构
type UserEvent struct {
	OpenId       string                 `json:"open_id"`         // 用户 OpenId
	EventType    string                 `json:"event_type"`      // 事件类型
	PlatformCode string                 `json:"platform_code"`   // 平台代码
	Timestamp    time.Time              `json:"timestamp"`       // 事件时间戳
	Message      string                 `json:"message"`         // 事件描述
	Extra        map[string]interface{} `json:"extra,omitempty"` // 额外信息
}

// UserKickOffEvent 用户踢下线事件结构
type UserKickOffEvent struct {
	OpenId       string    `json:"open_id"`          // 用户 OpenId
	EventType    string    `json:"event_type"`       // 事件类型
	PlatformCode string    `json:"platform_code"`    // 平台代码
	Timestamp    time.Time `json:"timestamp"`        // 事件时间戳
	Message      string    `json:"message"`          // 事件描述
	Reason       string    `json:"reason,omitempty"` // 踢下线原因
}

// UserLoginEvent 用户登录事件结构
type UserLoginEvent struct {
	OpenId       string    `json:"open_id"`              // 用户 OpenId
	EventType    string    `json:"event_type"`           // 事件类型
	PlatformCode string    `json:"platform_code"`        // 平台代码
	Timestamp    time.Time `json:"timestamp"`            // 事件时间戳
	Message      string    `json:"message"`              // 事件描述
	ClientIP     string    `json:"client_ip,omitempty"`  // 客户端IP
	UserAgent    string    `json:"user_agent,omitempty"` // 用户代理
}

// UserLogoutEvent 用户退出事件结构
type UserLogoutEvent struct {
	OpenId       string    `json:"open_id"`            // 用户 OpenId
	EventType    string    `json:"event_type"`         // 事件类型
	PlatformCode string    `json:"platform_code"`      // 平台代码
	Timestamp    time.Time `json:"timestamp"`          // 事件时间戳
	Message      string    `json:"message"`            // 事件描述
	Duration     int64     `json:"duration,omitempty"` // 在线时长（秒）
}

// EventHandler 事件处理函数类型
type EventHandler func(payload string)

// KickOffEventHandler 踢下线事件处理函数类型
type KickOffEventHandler func(event UserKickOffEvent)

// LoginEventHandler 登录事件处理函数类型
type LoginEventHandler func(event UserLoginEvent)

// LogoutEventHandler 退出事件处理函数类型
type LogoutEventHandler func(event UserLogoutEvent)

// Config 客户端配置
type Config struct {
	RedisURL    string        // Redis 连接 URL
	Password    string        // Redis 密码
	DB          int           // Redis 数据库编号
	Timeout     time.Duration // 连接超时时间
	PoolSize    int           // 连接池大小
	MaxRetries  int           // 最大重试次数
	IdleTimeout time.Duration // 空闲连接超时时间
	Logger      Logger        // 自定义日志器
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DB:          DefaultRedisDB,
		Timeout:     DefaultTimeout * time.Second,
		PoolSize:    10,
		MaxRetries:  3,
		IdleTimeout: 5 * time.Minute,
	}
}

// Client 通知客户端接口
type Client interface {
	// 发布事件
	PublishKickOff(openId, platformCode string, reason ...string) error
	PublishLogin(openId, platformCode string, clientIP, userAgent string) error
	PublishLogout(openId, platformCode string, duration int64) error
	PublishCustomEvent(openId, eventType, platformCode, message string, extra map[string]interface{}) error

	// 订阅事件
	SubscribeKickOff(openId string, handler EventHandler) (string, error)
	SubscribeLogin(openId string, handler EventHandler) (string, error)
	SubscribeLogout(openId string, handler EventHandler) (string, error)
	SubscribeCustomEvent(channel string, handler EventHandler) (string, error)

	// 订阅事件（类型化处理器）
	SubscribeKickOffTyped(openId string, handler KickOffEventHandler) (string, error)
	SubscribeLoginTyped(openId string, handler LoginEventHandler) (string, error)
	SubscribeLogoutTyped(openId string, handler LogoutEventHandler) (string, error)

	// 批量订阅
	SubscribeMultipleKickOff(openIds []string, handler EventHandler) ([]string, error)

	// 取消订阅
	Unsubscribe(subscriptionId string) error
	UnsubscribeByChannel(channel string) error
	UnsubscribeAll() error
	GetActiveSubscriptions() []SubscriptionInfo

	// 连接管理
	Close() error
	Ping(ctx context.Context) error
	IsConnected() bool
}

// SubscriptionInfo 订阅信息
type SubscriptionInfo struct {
	ID      string `json:"id"`      // 订阅唯一ID
	Channel string `json:"channel"` // 订阅的频道
	OpenId  string `json:"open_id"` // 用户OpenId（如果适用）
	Active  bool   `json:"active"`  // 是否活跃
}
