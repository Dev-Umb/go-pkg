package util

import (
	"encoding/json"
	"fmt"
	"regexp"
	"unicode"
)

// RemoveSpecialChars 移除字符串中的特殊字符，只保留字母和数字
// 注意：对于中文字符，不会处理成拼音
func RemoveSpecialChars(s string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(s, "")
}

// ConvertToPinyin 将中文字符转换为拼音首字母缩写，非中文字符保持不变
// 由于目前没有引入拼音库，对中文字符直接使用字母替代
func ConvertToPinyin(s string) string {
	result := ""
	for _, runeValue := range s {
		if unicode.Is(unicode.Han, runeValue) {
			// 对于中文字符，使用字母"y"替代
			// 理想情况下，这里应该使用拼音库转换
			result += "y"
		} else if unicode.IsLetter(runeValue) || unicode.IsDigit(runeValue) {
			// 对于字母和数字，直接保留
			result += string(runeValue)
		}
	}
	return result
}

// JsonStrToMap 将JSON字符串转换为map
func JsonStrToMap(jsonStr string, result interface{}) error {
	if jsonStr == "" {
		return nil
	}

	// 使用标准库的json包进行转换
	err := json.Unmarshal([]byte(jsonStr), result)
	if err != nil {
		return fmt.Errorf("convert json string to map error: %w", err)
	}

	return nil
}
