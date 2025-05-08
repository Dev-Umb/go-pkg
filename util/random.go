/**
* @Email: i@umb.ink
 */
package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var (
	letters      = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	numbers      = []rune("0123456789")
	lettersChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = lettersChars[rand.Intn(len(lettersChars))]
	}
	return string(b)
}

// GenerateRandomCode 生成指定长度的随机数字验证码
func GenerateRandomCode(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = numbers[rand.Intn(len(numbers))]
	}
	return string(b)
}

// GenerateUserId 生成用户ID
func GenerateUserId() string {
	// 获取时间戳
	timestamp := time.Now().UnixNano() / 1000000 // 毫秒时间戳
	// 生成随机字符串
	randomStr := GenerateRandomString(8)
	// 组合用户ID
	return fmt.Sprintf("u%d%s", timestamp, randomStr)
}

// GenerateRandomPhoneNumber 生成随机手机号（仅用于测试）
func GenerateRandomPhoneNumber() string {
	// 中国手机号前三位
	prefixes := []string{"130", "131", "132", "133", "134", "135", "136", "137", "138", "139",
		"150", "151", "152", "153", "155", "156", "157", "158", "159",
		"170", "176", "177", "178",
		"180", "181", "182", "183", "184", "185", "186", "187", "188", "189"}

	// 随机选择前缀
	prefix := prefixes[rand.Intn(len(prefixes))]

	// 生成8位随机数字
	randomPart := GenerateRandomCode(8)

	return prefix + randomPart
}

// GenerateUUID 生成简单的UUID
func GenerateUUID() string {
	segments := make([]string, 5)
	segments[0] = GenerateRandomString(8)
	segments[1] = GenerateRandomString(4)
	segments[2] = GenerateRandomString(4)
	segments[3] = GenerateRandomString(4)
	segments[4] = GenerateRandomString(12)

	return strings.Join(segments, "-")
}
