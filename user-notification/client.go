package notification

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// subscription 订阅信息结构
type subscription struct {
	id      string
	channel string
	openId  string
	cancel  context.CancelFunc
	pubsub  *redis.PubSub
	active  bool
}

// client 实现 Client 接口
type client struct {
	rdb           *redis.Client
	config        *Config
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
	closed        bool
	logger        Logger
	subscriptions map[string]*subscription // 订阅管理
	subMu         sync.RWMutex             // 订阅操作锁
}

// Logger 日志接口，允许用户自定义日志实现
type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
}

// defaultLogger 默认日志实现
type defaultLogger struct{}

func (l *defaultLogger) Info(args ...interface{}) { log.Println("[INFO]", fmt.Sprint(args...)) }
func (l *defaultLogger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}
func (l *defaultLogger) Error(args ...interface{}) { log.Println("[ERROR]", fmt.Sprint(args...)) }
func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}
func (l *defaultLogger) Warn(args ...interface{}) { log.Println("[WARN]", fmt.Sprint(args...)) }
func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

// NewClient 创建新的通知客户端
func NewClient(redisURL string, options ...ClientOption) (Client, error) {
	config := DefaultConfig()
	config.RedisURL = redisURL

	// 应用选项
	for _, option := range options {
		option(config)
	}

	return NewClientWithConfig(config)
}

// NewClientWithConfig 使用配置创建客户端
func NewClientWithConfig(config *Config) (Client, error) {
	if config.RedisURL == "" {
		return nil, fmt.Errorf("redis URL is required")
	}

	// 解析 Redis URL
	opt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	// 应用配置
	if config.Password != "" {
		opt.Password = config.Password
	}
	opt.DB = config.DB
	opt.DialTimeout = config.Timeout
	opt.PoolSize = config.PoolSize
	opt.MaxRetries = config.MaxRetries
	opt.IdleTimeout = config.IdleTimeout

	// 创建 Redis 客户端
	rdb := redis.NewClient(opt)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// 创建客户端
	ctx, cancel = context.WithCancel(context.Background())
	c := &client{
		rdb:           rdb,
		config:        config,
		ctx:           ctx,
		cancel:        cancel,
		logger:        &defaultLogger{},
		subscriptions: make(map[string]*subscription),
	}

	// 如果配置中提供了自定义日志器，则使用它
	if config.Logger != nil {
		c.logger = config.Logger
	}

	return c, nil
}

// ClientOption 客户端选项
type ClientOption func(*Config)

// WithPassword 设置 Redis 密码
func WithPassword(password string) ClientOption {
	return func(c *Config) {
		c.Password = password
	}
}

// WithDB 设置 Redis 数据库编号
func WithDB(db int) ClientOption {
	return func(c *Config) {
		c.DB = db
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithPoolSize 设置连接池大小
func WithPoolSize(size int) ClientOption {
	return func(c *Config) {
		c.PoolSize = size
	}
}

// WithLogger 设置自定义日志器
func WithLogger(logger Logger) ClientOption {
	return func(c *Config) {
		c.Logger = logger
	}
}

// PublishKickOff 发布踢下线事件
func (c *client) PublishKickOff(openId, platformCode string, reason ...string) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	reasonText := "用户在其他设备登录，被踢下线"
	if len(reason) > 0 && reason[0] != "" {
		reasonText = reason[0]
	}

	event := UserKickOffEvent{
		OpenId:       openId,
		EventType:    EventTypeKickOff,
		PlatformCode: platformCode,
		Timestamp:    time.Now(),
		Message:      reasonText,
		Reason:       reasonText,
	}

	return c.publishEvent(RedisChannelUserKickOffPrefix+openId, event)
}

// PublishLogin 发布登录事件
func (c *client) PublishLogin(openId, platformCode string, clientIP, userAgent string) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	event := UserLoginEvent{
		OpenId:       openId,
		EventType:    EventTypeLogin,
		PlatformCode: platformCode,
		Timestamp:    time.Now(),
		Message:      "用户登录",
		ClientIP:     clientIP,
		UserAgent:    userAgent,
	}

	return c.publishEvent(RedisChannelUserLoginPrefix+openId, event)
}

// PublishLogout 发布退出事件
func (c *client) PublishLogout(openId, platformCode string, duration int64) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	event := UserLogoutEvent{
		OpenId:       openId,
		EventType:    EventTypeLogout,
		PlatformCode: platformCode,
		Timestamp:    time.Now(),
		Message:      "用户退出",
		Duration:     duration,
	}

	return c.publishEvent(RedisChannelUserLogoutPrefix+openId, event)
}

// PublishCustomEvent 发布自定义事件
func (c *client) PublishCustomEvent(openId, eventType, platformCode, message string, extra map[string]interface{}) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	event := UserEvent{
		OpenId:       openId,
		EventType:    eventType,
		PlatformCode: platformCode,
		Timestamp:    time.Now(),
		Message:      message,
		Extra:        extra,
	}

	// 根据事件类型选择频道
	var channel string
	switch eventType {
	case EventTypeKickOff:
		channel = RedisChannelUserKickOffPrefix + openId
	case EventTypeLogin:
		channel = RedisChannelUserLoginPrefix + openId
	case EventTypeLogout:
		channel = RedisChannelUserLogoutPrefix + openId
	default:
		channel = fmt.Sprintf("user:%s:%s", eventType, openId)
	}

	return c.publishEvent(channel, event)
}

// publishEvent 发布事件的通用方法
func (c *client) publishEvent(channel string, event interface{}) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		c.logger.Errorf("序列化事件失败: %v", err)
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = c.rdb.Publish(c.ctx, channel, string(eventData)).Err()
	if err != nil {
		c.logger.Errorf("发布事件到 Redis 失败: %v", err)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	c.logger.Infof("成功发布事件到频道: %s", channel)
	return nil
}

// generateSubscriptionID 生成唯一的订阅ID
func generateSubscriptionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// SubscribeKickOff 订阅踢下线事件
func (c *client) SubscribeKickOff(openId string, handler EventHandler) (string, error) {
	channel := RedisChannelUserKickOffPrefix + openId
	return c.SubscribeCustomEvent(channel, handler)
}

// SubscribeLogin 订阅登录事件
func (c *client) SubscribeLogin(openId string, handler EventHandler) (string, error) {
	channel := RedisChannelUserLoginPrefix + openId
	return c.SubscribeCustomEvent(channel, handler)
}

// SubscribeLogout 订阅退出事件
func (c *client) SubscribeLogout(openId string, handler EventHandler) (string, error) {
	channel := RedisChannelUserLogoutPrefix + openId
	return c.SubscribeCustomEvent(channel, handler)
}

// SubscribeCustomEvent 订阅自定义事件
func (c *client) SubscribeCustomEvent(channel string, handler EventHandler) (string, error) {
	if err := c.checkClosed(); err != nil {
		return "", err
	}

	// 生成订阅ID
	subscriptionID := generateSubscriptionID()

	// 创建订阅专用的上下文
	subCtx, subCancel := context.WithCancel(c.ctx)

	// 创建Redis订阅
	pubsub := c.rdb.Subscribe(subCtx, channel)

	// 创建订阅信息
	sub := &subscription{
		id:      subscriptionID,
		channel: channel,
		cancel:  subCancel,
		pubsub:  pubsub,
		active:  true,
	}

	// 如果是用户相关频道，提取OpenId
	if len(channel) > len(RedisChannelUserKickOffPrefix) {
		if channel[:len(RedisChannelUserKickOffPrefix)] == RedisChannelUserKickOffPrefix {
			sub.openId = channel[len(RedisChannelUserKickOffPrefix):]
		} else if channel[:len(RedisChannelUserLoginPrefix)] == RedisChannelUserLoginPrefix {
			sub.openId = channel[len(RedisChannelUserLoginPrefix):]
		} else if channel[:len(RedisChannelUserLogoutPrefix)] == RedisChannelUserLogoutPrefix {
			sub.openId = channel[len(RedisChannelUserLogoutPrefix):]
		}
	}

	// 保存订阅信息
	c.subMu.Lock()
	c.subscriptions[subscriptionID] = sub
	c.subMu.Unlock()

	// 启动订阅处理goroutine
	go func() {
		defer func() {
			// 清理资源
			pubsub.Close()
			c.subMu.Lock()
			if sub, exists := c.subscriptions[subscriptionID]; exists {
				sub.active = false
			}
			c.subMu.Unlock()
		}()

		c.logger.Infof("开始订阅频道: %s (ID: %s)", channel, subscriptionID)

		ch := pubsub.Channel()
		for {
			select {
			case msg := <-ch:
				if msg != nil {
					c.logger.Infof("接收到频道 %s 的消息: %s", msg.Channel, msg.Payload)
					if handler != nil {
						handler(msg.Payload)
					}
				}
			case <-subCtx.Done():
				c.logger.Infof("停止订阅频道: %s (ID: %s)", channel, subscriptionID)
				return
			}
		}
	}()

	return subscriptionID, nil
}

// SubscribeKickOffTyped 订阅踢下线事件（类型化处理器）
func (c *client) SubscribeKickOffTyped(openId string, handler KickOffEventHandler) (string, error) {
	return c.SubscribeKickOff(openId, func(payload string) {
		var event UserKickOffEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			c.logger.Errorf("解析踢下线事件失败: %v", err)
			return
		}
		handler(event)
	})
}

// SubscribeLoginTyped 订阅登录事件（类型化处理器）
func (c *client) SubscribeLoginTyped(openId string, handler LoginEventHandler) (string, error) {
	return c.SubscribeLogin(openId, func(payload string) {
		var event UserLoginEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			c.logger.Errorf("解析登录事件失败: %v", err)
			return
		}
		handler(event)
	})
}

// SubscribeLogoutTyped 订阅退出事件（类型化处理器）
func (c *client) SubscribeLogoutTyped(openId string, handler LogoutEventHandler) (string, error) {
	return c.SubscribeLogout(openId, func(payload string) {
		var event UserLogoutEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			c.logger.Errorf("解析退出事件失败: %v", err)
			return
		}
		handler(event)
	})
}

// SubscribeMultipleKickOff 批量订阅踢下线事件
func (c *client) SubscribeMultipleKickOff(openIds []string, handler EventHandler) ([]string, error) {
	var subscriptionIds []string
	for _, openId := range openIds {
		subId, err := c.SubscribeKickOff(openId, handler)
		if err != nil {
			// 如果出错，取消已经创建的订阅
			for _, id := range subscriptionIds {
				c.Unsubscribe(id)
			}
			return nil, fmt.Errorf("failed to subscribe kick off for user %s: %w", openId, err)
		}
		subscriptionIds = append(subscriptionIds, subId)
	}
	return subscriptionIds, nil
}

// Unsubscribe 取消指定的订阅
func (c *client) Unsubscribe(subscriptionId string) error {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	sub, exists := c.subscriptions[subscriptionId]
	if !exists {
		return fmt.Errorf("订阅ID %s 不存在", subscriptionId)
	}

	if !sub.active {
		return fmt.Errorf("订阅ID %s 已经不活跃", subscriptionId)
	}

	// 取消订阅
	sub.cancel()
	sub.active = false

	c.logger.Infof("已取消订阅: %s (频道: %s)", subscriptionId, sub.channel)
	return nil
}

// UnsubscribeByChannel 取消指定频道的所有订阅
func (c *client) UnsubscribeByChannel(channel string) error {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	var canceledCount int
	for _, sub := range c.subscriptions {
		if sub.channel == channel && sub.active {
			sub.cancel()
			sub.active = false
			canceledCount++
		}
	}

	if canceledCount == 0 {
		return fmt.Errorf("频道 %s 没有活跃的订阅", channel)
	}

	c.logger.Infof("已取消频道 %s 的 %d 个订阅", channel, canceledCount)
	return nil
}

// UnsubscribeAll 取消所有订阅
func (c *client) UnsubscribeAll() error {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	var canceledCount int
	for _, sub := range c.subscriptions {
		if sub.active {
			sub.cancel()
			sub.active = false
			canceledCount++
		}
	}

	c.logger.Infof("已取消所有订阅，总计 %d 个", canceledCount)
	return nil
}

// GetActiveSubscriptions 获取所有活跃的订阅信息
func (c *client) GetActiveSubscriptions() []SubscriptionInfo {
	c.subMu.RLock()
	defer c.subMu.RUnlock()

	var subscriptions []SubscriptionInfo
	for _, sub := range c.subscriptions {
		if sub.active {
			subscriptions = append(subscriptions, SubscriptionInfo{
				ID:      sub.id,
				Channel: sub.channel,
				OpenId:  sub.openId,
				Active:  sub.active,
			})
		}
	}

	return subscriptions
}

// Close 关闭客户端
func (c *client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	// 取消所有订阅
	c.UnsubscribeAll()

	c.cancel()
	if err := c.rdb.Close(); err != nil {
		return fmt.Errorf("failed to close redis client: %w", err)
	}

	c.closed = true
	c.logger.Info("客户端已关闭")
	return nil
}

// Ping 测试连接
func (c *client) Ping(ctx context.Context) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	return c.rdb.Ping(ctx).Err()
}

// IsConnected 检查是否已连接
func (c *client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return c.rdb.Ping(ctx).Err() == nil
}

// checkClosed 检查客户端是否已关闭
func (c *client) checkClosed() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}
	return nil
}
