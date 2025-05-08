package nacos_sdk

import (
	"github.com/Dev-Umb/go-pkg/logger"
	"github.com/Dev-Umb/go-pkg/util"
	"strconv"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// RegisterNacosService 注册服务到Nacos，并订阅相关服务
// 返回服务发现客户端，可用于后续操作
func RegisterNacosService() *naming_client.INamingClient {
	// 获取服务发现客户端
	namingClient, err := GetNamingClient()
	if err != nil {
		logger.Errorf("获取Nacos服务发现客户端失败: %v", err)
		return nil
	}

	// 获取本机IP地址
	ip := util.GetLocalIP()
	// 服务名称和分组
	serviceName := projectName
	serviceGroup := nacosGroup

	// 注册RPC服务
	portUint64, err := strconv.ParseUint(rpcPort, 10, 64)
	if err != nil {
		logger.Fatalf("解析RPCPort失败: %v", err)
	}

	// 注册RPC服务
	metadata := map[string]string{"version": "1.0.0"}
	success, err := RegisterServiceInstance(serviceName, ip, portUint64, serviceGroup, metadata)
	if err != nil {
		logger.Warnf("注册RPC服务失败: %v", err)
	} else if success {
		logger.Infof("成功注册RPC服务: %s, 分组: %s, IP: %s, 端口: %d", serviceName, serviceGroup, ip, portUint64)
		// 也订阅自己，便于监控
		err = SubscribeService(serviceName, serviceGroup, func(instances []model.Instance, err error) {
			if err != nil {
				logger.Warnf("服务订阅回调错误: %v", err)
				return
			}
			if len(instances) > 0 {
				logger.Infof("服务 %s 实例发生变化，当前实例数: %d", serviceName, len(instances))
				for i, instance := range instances {
					logger.Debugf("实例 %d: %s:%d, 健康状态: %v", i+1, instance.Ip, instance.Port, instance.Healthy)
				}
			} else {
				logger.Warnf("服务 %s 当前没有可用实例", serviceName)
			}
		})
		if err != nil {
			logger.Warnf("订阅服务失败: %v", err)
		} else {
			logger.Infof("成功订阅服务: %s, 分组: %s", serviceName, serviceGroup)
		}
	}

	if namingClient == nil {
		logger.Errorf("RegisterNacos Error! namingClient is nil")
		return nil
	}

	// 等待一段时间，确保服务注册完成
	time.Sleep(1 * time.Second)
	namingClientPtr := &namingClient
	return namingClientPtr
}

// 验证服务注册状态
func verifyRegisteredServices(serviceName, wsServiceName, serviceGroup string) {
	client, err := GetNamingClient()
	if err != nil {
		logger.Errorf("获取Nacos客户端失败: %v", err)
		return
	}

	// 获取RPC服务实例
	service, err := client.GetService(vo.GetServiceParam{
		ServiceName: serviceName,
		GroupName:   serviceGroup,
	})
	if err != nil {
		logger.Errorf("获取RPC服务列表失败: %+v", err)
	} else {
		if len(service.Hosts) > 0 {
			logger.Infof("RPC服务 %s 实例数量: %d", serviceName, len(service.Hosts))
			for i, instance := range service.Hosts {
				logger.Infof("实例 %d: %s:%d, 健康状态: %v, 元数据: %v",
					i+1, instance.Ip, instance.Port, instance.Healthy, instance.Metadata)
			}
		} else {
			logger.Warnf("RPC服务 %s 当前没有实例，可能注册未生效", serviceName)
		}
	}

	// 获取WebSocket服务实例
	wsService, err := client.GetService(vo.GetServiceParam{
		ServiceName: wsServiceName,
		GroupName:   serviceGroup,
	})
	if err != nil {
		logger.Errorf("获取WebSocket服务列表失败: %+v", err)
	} else {
		if len(wsService.Hosts) > 0 {
			logger.Infof("WebSocket服务 %s 实例数量: %d", wsServiceName, len(wsService.Hosts))
			for i, instance := range wsService.Hosts {
				logger.Infof("实例 %d: %s:%d, 健康状态: %v, 元数据: %v",
					i+1, instance.Ip, instance.Port, instance.Healthy, instance.Metadata)
			}
		} else {
			logger.Warnf("WebSocket服务 %s 当前没有实例，可能注册未生效", wsServiceName)
		}
	}

	// 通过SelectInstances再次尝试获取RPC实例
	instances, err := client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   serviceGroup,
		HealthyOnly: false, // 不限制只查询健康实例，以便看到所有状态的实例
	})
	if err != nil {
		logger.Errorf("SelectInstances RPC服务失败: %+v", err)
	} else {
		logger.Infof("SelectInstances获取到RPC实例数量: %d", len(instances))
		for i, ins := range instances {
			logger.Infof("实例 %d: %s:%d, 健康: %v, 启用: %v",
				i+1, ins.Ip, ins.Port, ins.Healthy, ins.Enable)
		}
	}

	// 通过SelectInstances再次尝试获取WebSocket实例
	wsInstances, err := client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: wsServiceName,
		GroupName:   serviceGroup,
		HealthyOnly: false,
	})
	if err != nil {
		logger.Errorf("SelectInstances WebSocket服务失败: %+v", err)
	} else {
		logger.Infof("SelectInstances获取到WebSocket实例数量: %d", len(wsInstances))
		for i, ins := range wsInstances {
			logger.Infof("WebSocket实例 %d: %s:%d, 健康: %v, 启用: %v",
				i+1, ins.Ip, ins.Port, ins.Healthy, ins.Enable)
		}
	}
}
