// Package notification provides user account notification functionality via Redis pub/sub
/**
* @Author: chenhao29
* @Date: 2024/6/10
* @QQ: 1149558764
* @Email: i@umb.ink
 */
package notification

// Redis 频道前缀模板
const (
	// RedisChannelUserKickOffPrefix 用户被踢下线通知频道前缀，后接用户 OpenID
	RedisChannelUserKickOffPrefix = "user:kickoff:"
	// RedisChannelUserLoginPrefix 用户登录通知频道前缀，后接用户 OpenID
	RedisChannelUserLoginPrefix = "user:login:"
	// RedisChannelUserLogoutPrefix 用户退出通知频道前缀，后接用户 OpenID
	RedisChannelUserLogoutPrefix = "user:logout:"
)

// 事件类型常量
const (
	// EventTypeKickOff 踢下线事件
	EventTypeKickOff = "kick_off"
	// EventTypeLogin 登录事件
	EventTypeLogin = "login"
	// EventTypeLogout 退出事件
	EventTypeLogout = "logout"
	// EventTypeForceOffline 强制下线事件
	EventTypeForceOffline = "force_offline"
)

// 默认配置
const (
	// DefaultRedisDB 默认 Redis 数据库编号
	DefaultRedisDB = 0
	// DefaultTimeout 默认超时时间（秒）
	DefaultTimeout = 5
)
