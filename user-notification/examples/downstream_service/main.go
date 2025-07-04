// 下游服务示例：展示如何订阅用户账户事件并进行处理
package main

import (
	notification "github.com/Dev-Umb/go-pkg/user-notification"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// ============ 初始化全局通知客户端 ============
	log.Println("初始化下游服务...")

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

	// ============ 启动事件订阅 ============
	log.Println("启动事件订阅服务...")

	// 订阅需要监听的用户列表（实际应用中可能从数据库获取）
	userIds := []string{
		"user_openid_123456",
		"user_openid_789012",
		"user_openid_345678",
	}

	// 启动订阅服务
	startEventSubscription(userIds)

	// ============ 模拟发布一些事件用于测试 ============
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("模拟发布测试事件...")

		// 模拟用户被踢下线
		notification.PublishKickOff("user_openid_123456", "game_platform_001", "在其他设备登录")

		// 模拟用户登录
		notification.PublishLogin("user_openid_789012", "web_platform_002", "192.168.1.100", "Mozilla/5.0")

		// 模拟用户退出
		notification.PublishLogout("user_openid_345678", "mobile_platform_003", 7200)
	}()

	// ============ 等待退出信号 ============
	log.Println("下游服务已启动，等待事件...")

	// 设置信号处理，优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("接收到退出信号，正在关闭服务...")
}

// startEventSubscription 启动事件订阅服务
func startEventSubscription(userIds []string) {
	log.Printf("开始订阅 %d 个用户的事件", len(userIds))

	// 为每个用户订阅踢下线事件
	for _, userId := range userIds {
		// 订阅踢下线事件
		err := notification.SubscribeKickOffTyped(userId, func(event notification.UserKickOffEvent) {
			handleUserKickOff(event)
		})
		if err != nil {
			log.Printf("订阅用户 %s 踢下线事件失败: %v", userId, err)
			continue
		}

		// 订阅登录事件
		err = notification.SubscribeLoginTyped(userId, func(event notification.UserLoginEvent) {
			handleUserLogin(event)
		})
		if err != nil {
			log.Printf("订阅用户 %s 登录事件失败: %v", userId, err)
			continue
		}

		// 订阅退出事件
		err = notification.SubscribeLogoutTyped(userId, func(event notification.UserLogoutEvent) {
			handleUserLogout(event)
		})
		if err != nil {
			log.Printf("订阅用户 %s 退出事件失败: %v", userId, err)
			continue
		}

		log.Printf("成功订阅用户 %s 的所有事件", userId)
	}

	log.Println("所有事件订阅完成")
}

// handleUserKickOff 处理用户踢下线事件
func handleUserKickOff(event notification.UserKickOffEvent) {
	log.Printf("🚨 [踢下线事件] 用户: %s, 平台: %s, 原因: %s, 时间: %v",
		event.OpenId, event.PlatformCode, event.Reason, event.Timestamp)

	// 实际处理逻辑
	go func() {
		// 1. 通知前端用户下线
		notifyFrontend(event.OpenId, "kick_off", event.Reason)

		// 2. 清理用户相关缓存
		clearUserCache(event.OpenId, event.PlatformCode)

		// 3. 记录审计日志
		logAuditEvent("user_kick_off", event.OpenId, event.PlatformCode, event.Reason)

		// 4. 发送通知给运营团队（如果需要）
		if shouldNotifyOps(event.PlatformCode) {
			notifyOperations(event.OpenId, event.PlatformCode, "用户被踢下线")
		}

		log.Printf("✅ 踢下线事件处理完成: %s", event.OpenId)
	}()
}

// handleUserLogin 处理用户登录事件
func handleUserLogin(event notification.UserLoginEvent) {
	log.Printf("🔐 [登录事件] 用户: %s, 平台: %s, IP: %s, 时间: %v",
		event.OpenId, event.PlatformCode, event.ClientIP, event.Timestamp)

	// 实际处理逻辑
	go func() {
		// 1. 更新用户在线状态
		updateUserOnlineStatus(event.OpenId, true)

		// 2. 记录登录日志
		logUserActivity("login", event.OpenId, event.PlatformCode, event.ClientIP)

		// 3. 检查异常登录（例如：异地登录）
		checkAbnormalLogin(event.OpenId, event.ClientIP, event.UserAgent)

		// 4. 发送欢迎消息
		sendWelcomeMessage(event.OpenId, event.PlatformCode)

		log.Printf("✅ 登录事件处理完成: %s", event.OpenId)
	}()
}

// handleUserLogout 处理用户退出事件
func handleUserLogout(event notification.UserLogoutEvent) {
	log.Printf("🚪 [退出事件] 用户: %s, 平台: %s, 在线时长: %d秒, 时间: %v",
		event.OpenId, event.PlatformCode, event.Duration, event.Timestamp)

	// 实际处理逻辑
	go func() {
		// 1. 更新用户离线状态
		updateUserOnlineStatus(event.OpenId, false)

		// 2. 记录退出日志和在线时长
		logUserActivity("logout", event.OpenId, event.PlatformCode, "")
		logUserOnlineTime(event.OpenId, event.Duration)

		// 3. 清理用户会话数据
		clearUserSession(event.OpenId, event.PlatformCode)

		// 4. 计算用户活跃度分数
		calculateUserActivityScore(event.OpenId, event.Duration)

		log.Printf("✅ 退出事件处理完成: %s", event.OpenId)
	}()
}

// ============ 模拟的业务处理函数 ============

func notifyFrontend(openId, eventType, message string) {
	log.Printf("📱 通知前端: 用户 %s, 事件 %s, 消息: %s", openId, eventType, message)
	// 实际实现：通过 WebSocket 或 Server-Sent Events 通知前端
}

func clearUserCache(openId, platformCode string) {
	log.Printf("🗑️  清理缓存: 用户 %s, 平台 %s", openId, platformCode)
	// 实际实现：清理 Redis 缓存、内存缓存等
}

func logAuditEvent(eventType, openId, platformCode, details string) {
	log.Printf("📝 审计日志: %s - 用户 %s, 平台 %s, 详情: %s", eventType, openId, platformCode, details)
	// 实际实现：写入审计日志表
}

func shouldNotifyOps(platformCode string) bool {
	// 实际实现：根据平台重要性决定是否通知运营
	return platformCode == "game_platform_001"
}

func notifyOperations(openId, platformCode, message string) {
	log.Printf("📢 通知运营: 用户 %s, 平台 %s, 消息: %s", openId, platformCode, message)
	// 实际实现：发送邮件、短信或推送消息给运营团队
}

func updateUserOnlineStatus(openId string, isOnline bool) {
	status := "离线"
	if isOnline {
		status = "在线"
	}
	log.Printf("🔄 更新用户状态: %s -> %s", openId, status)
	// 实际实现：更新数据库中的用户状态
}

func logUserActivity(activity, openId, platformCode, clientIP string) {
	log.Printf("📊 记录用户活动: %s - 用户 %s, 平台 %s, IP: %s", activity, openId, platformCode, clientIP)
	// 实际实现：记录用户活动日志
}

func checkAbnormalLogin(openId, clientIP, userAgent string) {
	log.Printf("🔍 检查异常登录: 用户 %s, IP: %s", openId, clientIP)
	// 实际实现：检查IP地理位置、设备指纹等
}

func sendWelcomeMessage(openId, platformCode string) {
	log.Printf("👋 发送欢迎消息: 用户 %s, 平台 %s", openId, platformCode)
	// 实际实现：发送欢迎消息或推送
}

func logUserOnlineTime(openId string, duration int64) {
	hours := duration / 3600
	minutes := (duration % 3600) / 60
	log.Printf("⏱️  记录在线时长: 用户 %s, %d小时%d分钟", openId, hours, minutes)
	// 实际实现：记录用户在线时长统计
}

func clearUserSession(openId, platformCode string) {
	log.Printf("🧹 清理用户会话: 用户 %s, 平台 %s", openId, platformCode)
	// 实际实现：清理用户会话数据、临时文件等
}

func calculateUserActivityScore(openId string, duration int64) {
	score := duration / 60 // 简单的分数计算：每分钟1分
	log.Printf("🎯 计算活跃度分数: 用户 %s, 本次得分: %d", openId, score)
	// 实际实现：更新用户活跃度分数
}
