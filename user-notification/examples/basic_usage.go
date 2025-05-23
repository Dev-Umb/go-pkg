package main

import (
	"context"
	"fmt"
	"log"
	"time"

	notification "github.com/Dev-Umb/go-pkg/user-notification"
)

func main() {
	// 创建通知客户端
	client, err := notification.NewClient(
		"redis://localhost:6379",
		notification.WithPassword(""),           // Redis 密码（如果有）
		notification.WithDB(0),                  // Redis 数据库编号
		notification.WithTimeout(5*time.Second), // 连接超时
		notification.WithPoolSize(10),           // 连接池大小
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}
	log.Println("成功连接到 Redis")

	// 示例用户信息
	openId := "user_openid_123456"
	platformCode := "game_platform_001"

	// 1. 发布踢下线事件
	fmt.Println("\n=== 发布踢下线事件 ===")
	err = client.PublishKickOff(openId, platformCode, "用户在其他设备登录")
	if err != nil {
		log.Printf("发布踢下线事件失败: %v", err)
	} else {
		log.Println("成功发布踢下线事件")
	}

	// 2. 发布登录事件
	fmt.Println("\n=== 发布登录事件 ===")
	err = client.PublishLogin(openId, platformCode, "192.168.1.100", "Mozilla/5.0...")
	if err != nil {
		log.Printf("发布登录事件失败: %v", err)
	} else {
		log.Println("成功发布登录事件")
	}

	// 3. 发布退出事件
	fmt.Println("\n=== 发布退出事件 ===")
	err = client.PublishLogout(openId, platformCode, 3600) // 在线1小时
	if err != nil {
		log.Printf("发布退出事件失败: %v", err)
	} else {
		log.Println("成功发布退出事件")
	}

	// 4. 订阅踢下线事件（原始处理器）
	fmt.Println("\n=== 订阅踢下线事件 ===")
	kickOffSubId, err := client.SubscribeKickOff(openId, func(payload string) {
		log.Printf("接收到踢下线事件: %s", payload)
	})
	if err != nil {
		log.Printf("订阅踢下线事件失败: %v", err)
	} else {
		log.Printf("踢下线事件订阅ID: %s", kickOffSubId)
	}

	// 5. 订阅踢下线事件（类型化处理器）
	kickOffTypedSubId, err := client.SubscribeKickOffTyped(openId, func(event notification.UserKickOffEvent) {
		log.Printf("用户被踢下线: OpenId=%s, Platform=%s, Reason=%s, Time=%v",
			event.OpenId, event.PlatformCode, event.Reason, event.Timestamp)
	})
	if err != nil {
		log.Printf("订阅踢下线事件失败: %v", err)
	} else {
		log.Printf("踢下线事件类型化订阅ID: %s", kickOffTypedSubId)
	}

	// 6. 订阅登录事件（类型化处理器）
	fmt.Println("\n=== 订阅登录事件 ===")
	loginSubId, err := client.SubscribeLoginTyped(openId, func(event notification.UserLoginEvent) {
		log.Printf("用户登录: OpenId=%s, Platform=%s, ClientIP=%s, Time=%v",
			event.OpenId, event.PlatformCode, event.ClientIP, event.Timestamp)
	})
	if err != nil {
		log.Printf("订阅登录事件失败: %v", err)
	} else {
		log.Printf("登录事件订阅ID: %s", loginSubId)
	}

	// 7. 订阅退出事件（类型化处理器）
	fmt.Println("\n=== 订阅退出事件 ===")
	logoutSubId, err := client.SubscribeLogoutTyped(openId, func(event notification.UserLogoutEvent) {
		log.Printf("用户退出: OpenId=%s, Platform=%s, Duration=%d秒, Time=%v",
			event.OpenId, event.PlatformCode, event.Duration, event.Timestamp)
	})
	if err != nil {
		log.Printf("订阅退出事件失败: %v", err)
	} else {
		log.Printf("退出事件订阅ID: %s", logoutSubId)
	}

	// 8. 批量订阅多个用户的踢下线事件
	fmt.Println("\n=== 批量订阅踢下线事件 ===")
	userIds := []string{"user1", "user2", "user3"}
	batchSubIds, err := client.SubscribeMultipleKickOff(userIds, func(payload string) {
		log.Printf("批量订阅收到消息: %s", payload)
	})
	if err != nil {
		log.Printf("批量订阅失败: %v", err)
	} else {
		log.Printf("批量订阅ID: %v", batchSubIds)
	}

	// 9. 查看当前活跃订阅
	fmt.Println("\n=== 当前活跃订阅 ===")
	subscriptions := client.GetActiveSubscriptions()
	for i, sub := range subscriptions {
		log.Printf("%d. ID: %s, 频道: %s, 用户: %s", i+1, sub.ID, sub.Channel, sub.OpenId)
	}

	// 等待一段时间以接收消息
	fmt.Println("\n=== 等待接收消息 ===")
	time.Sleep(2 * time.Second)

	// 再次发布事件以测试订阅
	fmt.Println("\n=== 测试发布事件 ===")
	client.PublishKickOff(openId, platformCode, "测试踢下线")
	client.PublishLogin(openId, platformCode, "10.0.0.1", "TestAgent")
	client.PublishLogout(openId, platformCode, 1800)

	// 等待消息处理
	time.Sleep(2 * time.Second)

	// 10. 演示取消订阅
	fmt.Println("\n=== 取消订阅演示 ===")
	if loginSubId != "" {
		err = client.Unsubscribe(loginSubId)
		if err != nil {
			log.Printf("取消登录事件订阅失败: %v", err)
		} else {
			log.Printf("已取消登录事件订阅: %s", loginSubId)
		}
	}

	// 查看取消后的活跃订阅
	fmt.Println("\n=== 取消后的活跃订阅 ===")
	subscriptions = client.GetActiveSubscriptions()
	log.Printf("活跃订阅数量: %d", len(subscriptions))

	fmt.Println("\n=== 示例结束 ===")
}
