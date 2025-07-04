package nacos_sdk

import (
	"fmt"
	"log"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// 单例模式实现，确保整个应用只有一个nacos配置客户端实例
var (
	configClientInstance config_client.IConfigClient // 全局配置客户端实例
	once                 sync.Once                   // 用于确保初始化只执行一次
	initErr              error                       // 初始化过程中可能发生的错误
)

// GetConfigClient 获取单例配置客户端
// 第一次调用时会初始化客户端，后续调用返回已初始化的实例
// 返回配置客户端实例和可能的错误
func GetConfigClient() (config_client.IConfigClient, error) {
	once.Do(func() {
		// 初始化Nacos配置
		if err := initNacosConfig(); err != nil {
			initErr = fmt.Errorf("初始化Nacos配置失败: %v", err)
			return
		}

		// 创建配置客户端
		client, err := clients.CreateConfigClient(map[string]interface{}{
			"serverConfigs": serverConfigs,
			"clientConfig":  clientConfig,
		})
		if err != nil {
			initErr = fmt.Errorf("创建Nacos配置客户端失败: %v", err)
			return
		}
		configClientInstance = client
		log.Print("Nacos配置客户端初始化成功")
	})

	return configClientInstance, initErr
}

// GetConfigValue 获取配置值
// 根据dataId和group查询配置内容
// dataId: 配置ID
// group: 配置分组
// 返回配置内容和可能的错误
func GetConfigValue(dataId, group string) (string, error) {
	client, err := GetConfigClient()
	if err != nil {
		return "", err
	}

	value, err := client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {

		log.Printf("获取配置失败 [%s:%s]: %v", group, dataId, err)
		return "", err
	}

	return value, nil
}

// ListenConfigChange 监听配置变更
// 当配置发生变化时，会调用onChange回调函数
// dataId: 配置ID
// group: 配置分组
// onChange: 配置变更回调函数，参数为新的配置内容
// 返回可能的错误
func ListenConfigChange(dataId, group string, onChange func(data string)) error {
	client, err := GetConfigClient()
	if err != nil {
		return err
	}

	err = client.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("配置变更 [%s:%s]", group, dataId)
			onChange(data)
		},
	})
	if err != nil {
		log.Printf("监听配置失败 [%s:%s]: %v", group, dataId, err)
		return err
	}

	log.Printf("开始监听配置 [%s:%s]", group, dataId)
	return nil
}
