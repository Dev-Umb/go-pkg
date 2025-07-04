package main

import (
	"fmt"
	"log"
	"time"

	notification "github.com/Dev-Umb/go-pkg/user-notification"
)

func main() {
	// 初始化全局客户端
	err := notification.InitGlobalClient("redis://localhost:6379/0")
	if err != nil {
		log.Printf("初始化通知客户端失败: %v", err)
	}
	defer notification.CloseGlobalClient()

	// 示例用户ID
	userId := "user123"

	fmt.Println("=== 订阅取消示例 ===")

	// 1. 订阅踢下线事件并获取订阅ID
	fmt.Printf("订阅用户 %s 的踢下线事件...\n", userId)
	kickOffSubId, err := notification.SubscribeKickOff(userId, func(payload string) {
		fmt.Printf("收到踢下线事件: %s\n", payload)
	})
	if err != nil {
		log.Printf("订阅踢下线事件失败: %v", err)
	}
	fmt.Printf("踢下线事件订阅ID: %s\n", kickOffSubId)

	// 2. 订阅登录事件
	fmt.Printf("订阅用户 %s 的登录事件...\n", userId)
	loginSubId, err := notification.SubscribeLogin(userId, func(payload string) {
		fmt.Printf("收到登录事件: %s\n", payload)
	})
	if err != nil {
		log.Printf("订阅登录事件失败: %v", err)
	}
	fmt.Printf("登录事件订阅ID: %s\n", loginSubId)

	// 3. 订阅退出事件（使用类型化处理器）
	fmt.Printf("订阅用户 %s 的退出事件（类型化）...\n", userId)
	logoutSubId, err := notification.SubscribeLogoutTyped(userId, func(event notification.UserLogoutEvent) {
		fmt.Printf("收到退出事件: OpenId=%s, Duration=%d秒\n", event.OpenId, event.Duration)
	})
	if err != nil {
		log.Printf("订阅退出事件失败: %v", err)
	}
	fmt.Printf("退出事件订阅ID: %s\n", logoutSubId)

	// 4. 查看当前活跃的订阅
	fmt.Println("\n=== 当前活跃订阅 ===")
	subscriptions := notification.GetActiveSubscriptions()
	for i, sub := range subscriptions {
		fmt.Printf("%d. ID: %s, 频道: %s, 用户: %s\n", i+1, sub.ID, sub.Channel, sub.OpenId)
	}

	// 5. 模拟发布一些事件
	fmt.Println("\n=== 发布测试事件 ===")

	// 发布登录事件
	fmt.Println("发布登录事件...")
	err = notification.PublishLogin(userId, "web", "192.168.1.100", "Mozilla/5.0")
	if err != nil {
		log.Printf("发布登录事件失败: %v", err)
	}

	// 等待事件处理
	time.Sleep(1 * time.Second)

	// 发布踢下线事件
	fmt.Println("发布踢下线事件...")
	err = notification.PublishKickOff(userId, "mobile", "在其他设备登录")
	if err != nil {
		log.Printf("发布踢下线事件失败: %v", err)
	}

	// 等待事件处理
	time.Sleep(1 * time.Second)

	// 6. 取消单个订阅
	fmt.Printf("\n=== 取消登录事件订阅 (ID: %s) ===\n", loginSubId)
	err = notification.Unsubscribe(loginSubId)
	if err != nil {
		log.Printf("取消订阅失败: %v", err)
	} else {
		fmt.Println("登录事件订阅已取消")
	}

	// 7. 再次查看活跃订阅
	fmt.Println("\n=== 取消后的活跃订阅 ===")
	subscriptions = notification.GetActiveSubscriptions()
	for i, sub := range subscriptions {
		fmt.Printf("%d. ID: %s, 频道: %s, 用户: %s\n", i+1, sub.ID, sub.Channel, sub.OpenId)
	}

	// 8. 再次发布登录事件，验证已取消的订阅不会接收到
	fmt.Println("\n=== 验证取消效果 ===")
	fmt.Println("再次发布登录事件（应该不会收到通知）...")
	err = notification.PublishLogin(userId, "web", "192.168.1.101", "Chrome")
	if err != nil {
		log.Printf("发布登录事件失败: %v", err)
	}

	// 等待事件处理
	time.Sleep(1 * time.Second)

	// 9. 按频道取消订阅
	fmt.Printf("\n=== 取消用户 %s 的所有踢下线订阅 ===\n", userId)
	kickOffChannel := "user:kickoff:" + userId
	err = notification.UnsubscribeByChannel(kickOffChannel)
	if err != nil {
		log.Printf("按频道取消订阅失败: %v", err)
	} else {
		fmt.Printf("频道 %s 的所有订阅已取消\n", kickOffChannel)
	}

	// 10. 最终查看活跃订阅
	fmt.Println("\n=== 最终活跃订阅 ===")
	subscriptions = notification.GetActiveSubscriptions()
	if len(subscriptions) == 0 {
		fmt.Println("无活跃订阅")
	} else {
		for i, sub := range subscriptions {
			fmt.Printf("%d. ID: %s, 频道: %s, 用户: %s\n", i+1, sub.ID, sub.Channel, sub.OpenId)
		}
	}

	// 11. 发布退出事件测试剩余订阅
	fmt.Println("\n=== 测试剩余订阅 ===")
	fmt.Println("发布退出事件...")
	err = notification.PublishLogout(userId, "web", 3600) // 在线1小时
	if err != nil {
		log.Printf("发布退出事件失败: %v", err)
	}

	// 等待事件处理
	time.Sleep(1 * time.Second)

	// 12. 取消所有订阅
	fmt.Println("\n=== 取消所有剩余订阅 ===")
	err = notification.UnsubscribeAll()
	if err != nil {
		log.Printf("取消所有订阅失败: %v", err)
	} else {
		fmt.Println("所有订阅已取消")
	}

	// 最终验证
	fmt.Println("\n=== 最终验证 ===")
	subscriptions = notification.GetActiveSubscriptions()
	fmt.Printf("活跃订阅数量: %d\n", len(subscriptions))

	fmt.Println("\n示例执行完毕")
}
