// Package notification provides user account notification functionality via Redis pub/sub
/**
* @Author: chenhao29
* @Date: 2024/6/10
* @QQ: 1149558764
* @Email: i@umb.ink
 */
package notification

import (
	"fmt"
	"sync"
)

var (
	// globalClient 全局客户端实例
	globalClient Client
	// once 确保只初始化一次
	once sync.Once
	// mu 保护全局客户端的读写
	mu sync.RWMutex
	// initialized 标记是否已初始化
	initialized bool
)

// InitGlobalClient 初始化全局通知客户端
// 这个方法应该在应用启动时调用一次
func InitGlobalClient(redisURL string, options ...ClientOption) error {
	var initErr error
	once.Do(func() {
		client, err := NewClient(redisURL, options...)
		if err != nil {
			initErr = fmt.Errorf("初始化全局通知客户端失败: %w", err)
			return
		}

		mu.Lock()
		globalClient = client
		initialized = true
		mu.Unlock()
	})

	return initErr
}

// GetGlobalClient 获取全局通知客户端实例
func GetGlobalClient() Client {
	mu.RLock()
	defer mu.RUnlock()

	if !initialized || globalClient == nil {
		panic("全局通知客户端未初始化，请先调用 InitGlobalClient")
	}

	return globalClient
}

// IsGlobalClientInitialized 检查全局客户端是否已初始化
func IsGlobalClientInitialized() bool {
	mu.RLock()
	defer mu.RUnlock()

	return initialized && globalClient != nil
}

// CloseGlobalClient 关闭全局客户端
func CloseGlobalClient() error {
	mu.Lock()
	defer mu.Unlock()

	if !initialized || globalClient == nil {
		return nil
	}

	err := globalClient.Close()
	globalClient = nil
	initialized = false

	return err
}

// 全局便捷方法 - 发布事件

// PublishKickOff 发布踢下线事件（全局方法）
func PublishKickOff(openId, platformCode string, reason ...string) error {
	return GetGlobalClient().PublishKickOff(openId, platformCode, reason...)
}

// PublishLogin 发布登录事件（全局方法）
func PublishLogin(openId, platformCode string, clientIP, userAgent string) error {
	return GetGlobalClient().PublishLogin(openId, platformCode, clientIP, userAgent)
}

// PublishLogout 发布退出事件（全局方法）
func PublishLogout(openId, platformCode string, duration int64) error {
	return GetGlobalClient().PublishLogout(openId, platformCode, duration)
}

// PublishCustomEvent 发布自定义事件（全局方法）
func PublishCustomEvent(openId, eventType, platformCode, message string, extra map[string]interface{}) error {
	return GetGlobalClient().PublishCustomEvent(openId, eventType, platformCode, message, extra)
}

// 全局便捷方法 - 订阅事件

// SubscribeKickOff 订阅踢下线事件（全局方法）
func SubscribeKickOff(openId string, handler EventHandler) error {
	return GetGlobalClient().SubscribeKickOff(openId, handler)
}

// SubscribeLogin 订阅登录事件（全局方法）
func SubscribeLogin(openId string, handler EventHandler) error {
	return GetGlobalClient().SubscribeLogin(openId, handler)
}

// SubscribeLogout 订阅退出事件（全局方法）
func SubscribeLogout(openId string, handler EventHandler) error {
	return GetGlobalClient().SubscribeLogout(openId, handler)
}

// SubscribeKickOffTyped 订阅踢下线事件（全局方法，类型化处理器）
func SubscribeKickOffTyped(openId string, handler KickOffEventHandler) error {
	return GetGlobalClient().SubscribeKickOffTyped(openId, handler)
}

// SubscribeLoginTyped 订阅登录事件（全局方法，类型化处理器）
func SubscribeLoginTyped(openId string, handler LoginEventHandler) error {
	return GetGlobalClient().SubscribeLoginTyped(openId, handler)
}

// SubscribeLogoutTyped 订阅退出事件（全局方法，类型化处理器）
func SubscribeLogoutTyped(openId string, handler LogoutEventHandler) error {
	return GetGlobalClient().SubscribeLogoutTyped(openId, handler)
}

// SubscribeMultipleKickOff 批量订阅踢下线事件（全局方法）
func SubscribeMultipleKickOff(openIds []string, handler EventHandler) error {
	return GetGlobalClient().SubscribeMultipleKickOff(openIds, handler)
}
