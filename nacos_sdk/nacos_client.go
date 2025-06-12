package nacos_sdk

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Dev-Umb/go-pkg/logger"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GetGRPCClient 获取gRPC客户端实例，使用泛型以支持不同类型的客户端
// T 是gRPC客户端接口类型
// serviceName 是要连接的服务名称
// group 是服务所在的分组，默认为当前服务的分组
// newClientFunc 是创建新客户端的函数
func GetGRPCClient[T any](serviceName string, groupName string, newClientFunc func(conn *grpc.ClientConn) T) (T, error) {
	// 使用服务发现客户端获取服务实例
	instance, err := GetHealthyInstance(serviceName, groupName)
	if err != nil {
		return *new(T), fmt.Errorf("获取服务实例失败: %v", err)
	}

	// 创建gRPC连接
	return createGRPCClient(*instance, serviceName, newClientFunc)
}

// CreateGRPCClientWithInstance 使用给定的服务实例创建gRPC客户端
// instance 是服务实例
// serviceName 是服务名称（用于日志记录）
// newClientFunc 是创建新客户端的函数
func CreateGRPCClientWithInstance[T any](instance model.Instance, serviceName string, newClientFunc func(conn *grpc.ClientConn) T) (T, error) {
	return createGRPCClient(instance, serviceName, newClientFunc)
}

// createGRPCClient 创建gRPC客户端的内部实现
func createGRPCClient[T any](instance model.Instance, serviceName string, newClientFunc func(conn *grpc.ClientConn) T) (T, error) {
	// 创建gRPC连接
	addr := fmt.Sprintf("%s:%d", instance.Ip, instance.Port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("连接服务实例失败: %s, 地址: %s, 错误: %v", serviceName, addr, err)
		return *new(T), fmt.Errorf("连接服务实例失败: %s, 错误: %v", serviceName, err)
	}

	// 创建客户端
	client := newClientFunc(conn)
	return client, nil
}

// 初始化Nacos客户端配置
func initNacosConfig() error {
	logger.Info(context.Background(), "initNacosConfig")
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// 创建nacos目录（使用绝对路径）
	nacosDir = filepath.Join(workDir, "nacos_sdk")
	logDir = filepath.Join(nacosDir, "log")
	configDir := filepath.Join(cacheDir, "config")

	// 确保目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 创建测试配置文件
	testConfigPath := filepath.Join(configDir, fmt.Sprintf("test@@%s@@.", "DEFAULT_GROUP"))
	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		// 创建空文件
		f, err := os.Create(testConfigPath)
		if err != nil {
			log.Fatalf("nacos_sdk err %v", err)
			return err
		}
		f.Close()
	}

	// Nacos服务器地址
	serverConfigs = []constant.ServerConfig{
		*constant.NewServerConfig(
			nacosAddress, nacosPort, constant.WithScheme("http")),
	}
	// 客户端配置
	clientConfig = constant.ClientConfig{
		NamespaceId:          nacosNameSpace, // 如果不需要命名空间，可以留空
		TimeoutMs:            10000,
		NotLoadCacheAtStart:  true,
		LogDir:               logDir,
		LogLevel:             "debug",
		UpdateThreadNum:      5,        // 更新线程数
		UpdateCacheWhenEmpty: true,     // 当服务列表为空时更新缓存
		BeatInterval:         1 * 1000, // 心跳间隔，单位毫秒（调整为更频繁的心跳）
	}
	return nil
}
