package main

import (
	"log"
	"time"

	notification "login-server/pkg/user-notification"
)

func main() {
	// ============ 应用启动时初始化全局客户端 ============
	err := notification.InitGlobalClient(
		"redis://localhost:6379",
		notification.WithPassword(""),           // Redis 密码（如果有）
		notification.WithDB(0),                  // Redis 数据库编号
		notification.WithTimeout(5*time.Second), // 连接超时
		notification.WithPoolSize(10),           // 连接池大小
	)
	if err != nil {
		log.Printf("初始化全局通知客户端失败: %v", err)
	}
	log.Println("全局通知客户端初始化成功")

	// 确保程序退出时关闭客户端
	defer func() {
		if err := notification.CloseGlobalClient(); err != nil {
			log.Printf("关闭全局客户端失败: %v", err)
		} else {
			log.Println("全局客户端已关闭")
		}
	}()

	// ============ 检查初始化状态 ============
	if notification.IsGlobalClientInitialized() {
		log.Println("全局客户端已初始化")
	}

	// ============ 使用全局便捷方法发布事件 ============
	openId := "user_singleton_123"
	platformCode := "game_platform_001"

	// 发布踢下线事件
	log.Println("\n=== 使用全局方法发布踢下线事件 ===")
	err = notification.PublishKickOff(openId, platformCode, "用户在其他设备登录")
	if err != nil {
		log.Printf("发布踢下线事件失败: %v", err)
	} else {
		log.Println("成功发布踢下线事件")
	}

	// 发布登录事件
	log.Println("\n=== 使用全局方法发布登录事件 ===")
	err = notification.PublishLogin(openId, platformCode, "192.168.1.100", "Mozilla/5.0...")
	if err != nil {
		log.Printf("发布登录事件失败: %v", err)
	} else {
		log.Println("成功发布登录事件")
	}

	// 发布退出事件
	log.Println("\n=== 使用全局方法发布退出事件 ===")
	err = notification.PublishLogout(openId, platformCode, 3600)
	if err != nil {
		log.Printf("发布退出事件失败: %v", err)
	} else {
		log.Println("成功发布退出事件")
	}

	// ============ 使用全局便捷方法订阅事件 ============
	log.Println("\n=== 使用全局方法订阅事件 ===")

	// 订阅踢下线事件（类型化处理器）
	err = notification.SubscribeKickOffTyped(openId, func(event notification.UserKickOffEvent) {
		log.Printf("[全局订阅] 用户被踢下线: OpenId=%s, Platform=%s, Reason=%s, Time=%v",
			event.OpenId, event.PlatformCode, event.Reason, event.Timestamp)
	})
	if err != nil {
		log.Printf("订阅踢下线事件失败: %v", err)
	}

	// 订阅登录事件（类型化处理器）
	err = notification.SubscribeLoginTyped(openId, func(event notification.UserLoginEvent) {
		log.Printf("[全局订阅] 用户登录: OpenId=%s, Platform=%s, ClientIP=%s, Time=%v",
			event.OpenId, event.PlatformCode, event.ClientIP, event.Timestamp)
	})
	if err != nil {
		log.Printf("订阅登录事件失败: %v", err)
	}

	// 订阅退出事件（类型化处理器）
	err = notification.SubscribeLogoutTyped(openId, func(event notification.UserLogoutEvent) {
		log.Printf("[全局订阅] 用户退出: OpenId=%s, Platform=%s, Duration=%d秒, Time=%v",
			event.OpenId, event.PlatformCode, event.Duration, event.Timestamp)
	})
	if err != nil {
		log.Printf("订阅退出事件失败: %v", err)
	}

	// 批量订阅多个用户
	userIds := []string{"user1", "user2", "user3"}
	err = notification.SubscribeMultipleKickOff(userIds, func(payload string) {
		log.Printf("[批量订阅] 收到踢下线消息: %s", payload)
	})
	if err != nil {
		log.Printf("批量订阅失败: %v", err)
	}

	// ============ 测试事件发布和接收 ============
	log.Println("\n=== 等待1秒后测试发布事件 ===")
	time.Sleep(1 * time.Second)

	// 再次发布事件测试订阅
	notification.PublishKickOff(openId, platformCode, "测试踢下线")
	notification.PublishLogin(openId, platformCode, "10.0.0.1", "TestAgent")
	notification.PublishLogout(openId, platformCode, 1800)

	// 等待消息处理
	time.Sleep(2 * time.Second)

	// ============ 演示在不同模块中的使用 ============
	log.Println("\n=== 模拟在不同模块中使用 ===")
	simulateUserService()
	simulateGameService()

	log.Println("\n=== 示例结束 ===")
}

// simulateUserService 模拟用户服务模块使用全局客户端
func simulateUserService() {
	log.Println("用户服务模块: 发布用户登录事件")
	err := notification.PublishLogin("module_user_001", "user_service", "127.0.0.1", "UserServiceBot")
	if err != nil {
		log.Printf("用户服务模块发布登录事件失败: %v", err)
	}
}

// simulateGameService 模拟游戏服务模块使用全局客户端
func simulateGameService() {
	log.Println("游戏服务模块: 发布自定义游戏事件")
	extra := map[string]interface{}{
		"level":  10,
		"score":  1000,
		"reward": "gold_coin",
	}
	err := notification.PublishCustomEvent("module_user_001", "game_level_up", "game_service", "恭喜升级", extra)
	if err != nil {
		log.Printf("游戏服务模块发布事件失败: %v", err)
	}
}
