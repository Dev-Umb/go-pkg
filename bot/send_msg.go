package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// SendMsg 发送钉钉机器人消息
// content: 消息内容
func SendMsg(content string) error {
	// 构建请求体
	requestBody := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": "[后端服务]" + content,
		},
		"at": map[string]interface{}{
			"isAtAll": false,
		},
	}

	// 将请求体转换为JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("JSON编码失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", botUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d，响应: %s", resp.StatusCode, string(body))
	}

	return nil
}
