package nacos_sdk

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// 单例模式实现
var (
	namingClientInstance naming_client.INamingClient
	namingOnce           sync.Once
	namingInitErr        error
)

// GetNamingClient 获取单例服务发现客户端
// 返回全局唯一的命名服务客户端实例和可能的错误
func GetNamingClient() (naming_client.INamingClient, error) {
	namingOnce.Do(func() {
		// 初始化Nacos配置
		if err := initNacosConfig(); err != nil {
			namingInitErr = fmt.Errorf("初始化Nacos配置失败: %v", err)
			return
		}

		// 创建服务发现客户端
		client, err := clients.CreateNamingClient(map[string]interface{}{
			"serverConfigs": serverConfigs,
			"clientConfig":  clientConfig,
		})
		if err != nil {
			namingInitErr = fmt.Errorf("创建Nacos服务发现客户端失败: %v", err)
			return
		}

		// 等待客户端连接就绪
		for i := 0; i < 10; i++ {
			// 尝试获取服务列表，检查连接状态
			_, err := client.GetAllServicesInfo(vo.GetAllServiceInfoParam{
				PageNo:   1,
				PageSize: 10,
			})
			if err == nil {
				// 连接成功
				namingClientInstance = client
				log.Printf("Nacos服务发现客户端初始化成功")
				return
			}
			log.Printf("等待Nacos服务发现客户端连接就绪，重试次数: %d, 错误: %v", i+1, err)
			time.Sleep(1 * time.Second)
		}

		if namingClientInstance == nil {
			namingInitErr = fmt.Errorf("初始化Nacos服务发现客户端超时")
		}
	})

	return namingClientInstance, namingInitErr
}

// GetHealthyInstance 获取一个健康的服务实例
// serviceName: 服务名称
// group: 服务分组
// 返回服务实例和可能的错误
func GetHealthyInstance(serviceName, group string) (*model.Instance, error) {
	client, err := GetNamingClient()
	if err != nil {
		return nil, err
	}

	instance, err := client.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   group,
	})
	if err != nil {
		log.Printf("获取健康服务实例失败 [%s:%s]: %v", group, serviceName, err)
		return nil, err
	}

	return instance, nil
}

// GetAllInstances 获取指定服务的所有实例
// serviceName: 服务名称
// group: 服务分组
// onlyHealthy: 是否只返回健康实例
// 返回服务实例列表和可能的错误
func GetAllInstances(serviceName, group string, onlyHealthy bool) ([]model.Instance, error) {
	client, err := GetNamingClient()
	if err != nil {
		return nil, err
	}

	instances, err := client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   group,
		HealthyOnly: onlyHealthy,
	})
	if err != nil {
		log.Printf("获取服务实例列表失败 [%s:%s] (onlyHealthy=%v): %v",
			group, serviceName, onlyHealthy, err)
		return nil, err
	}

	return instances, nil
}

// SubscribeService 订阅服务变更
// serviceName: 服务名称
// group: 服务分组
// callback: 服务变更回调函数
// 返回可能的错误
func SubscribeService(serviceName, group string, callback func(instances []model.Instance, err error)) error {
	client, err := GetNamingClient()
	if err != nil {
		return err
	}

	err = client.Subscribe(&vo.SubscribeParam{
		ServiceName:       serviceName,
		GroupName:         group,
		SubscribeCallback: callback,
	})
	if err != nil {
		log.Printf("订阅服务失败 [%s:%s]: %v", group, serviceName, err)
		return err
	}

	log.Printf("成功订阅服务 [%s:%s]", group, serviceName)
	return nil
}

// UnsubscribeService 取消订阅服务变更
// serviceName: 服务名称
// group: 服务分组
// callback: 服务变更回调函数
// 返回可能的错误
func UnsubscribeService(serviceName, group string, callback func(instances []model.Instance, err error)) error {
	client, err := GetNamingClient()
	if err != nil {
		return err
	}

	err = client.Unsubscribe(&vo.SubscribeParam{
		ServiceName:       serviceName,
		GroupName:         group,
		SubscribeCallback: callback,
	})
	if err != nil {
		log.Printf("取消订阅服务失败 [%s:%s]: %v", group, serviceName, err)
		return err
	}

	log.Printf("成功取消订阅服务 [%s:%s]", group, serviceName)
	return nil
}

// RegisterServiceInstance 注册服务实例
// serviceName: 服务名称
// ip: 服务IP
// port: 服务端口
// group: 服务分组
// metadata: 服务元数据
// 返回注册是否成功和可能的错误
func RegisterServiceInstance(serviceName, ip string, port uint64, group string, metadata map[string]string) (bool, error) {
	client, err := GetNamingClient()
	if err != nil {
		return false, err
	}

	// 重试注册服务
	var success bool
	for i := 0; i < 3; i++ {
		success, err = client.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          ip,
			Port:        port,
			ServiceName: serviceName,
			Weight:      10,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			Metadata:    metadata,
			GroupName:   group,
		})
		if err == nil && success {
			// 注册成功，记录服务信息到全局变量
			serviceKey := fmt.Sprintf("%s-%s-%d-%s", serviceName, ip, port, group)
			serviceCheckMutex.Lock()
			registeredServices[serviceKey] = &RegisteredServiceInfo{
				ServiceName:   serviceName,
				IP:            ip,
				Port:          port,
				Group:         group,
				NamingClient:  nil, // 不再保存客户端实例
				LastHeartbeat: time.Now(),
			}
			serviceCheckMutex.Unlock()

			log.Printf("服务注册成功 [%s:%s] IP:%s Port:%d", group, serviceName, ip, port)
			return success, nil
		}
		log.Printf("注册服务失败，重试次数: %d, 错误: %v", i+1, err)
		time.Sleep(1 * time.Second)
	}

	return success, err
}

// DeregisterServiceInstance 注销服务实例
// serviceName: 服务名称
// ip: 服务IP
// port: 服务端口
// group: 服务分组
// 返回注销是否成功和可能的错误
func DeregisterServiceInstance(serviceName, ip string, port uint64, group string) (bool, error) {
	client, err := GetNamingClient()
	if err != nil {
		return false, err
	}

	// 从记录中移除服务
	serviceKey := fmt.Sprintf("%s-%s-%d-%s", serviceName, ip, port, group)
	serviceCheckMutex.Lock()
	delete(registeredServices, serviceKey)
	serviceCheckMutex.Unlock()

	success, err := client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   group,
	})
	if err != nil {
		log.Printf("注销服务实例失败 [%s:%s] IP:%s Port:%d: %v",
			group, serviceName, ip, port, err)
		return false, err
	}

	log.Printf("服务注销成功 [%s:%s] IP:%s Port:%d", group, serviceName, ip, port)
	return success, nil
}
