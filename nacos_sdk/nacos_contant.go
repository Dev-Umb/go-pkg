package nacos_sdk

import (
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"sync"
	"time"
)

// 全局变量，用于存储Nacos客户端配置
var (
	serverConfigs []constant.ServerConfig
	clientConfig  constant.ClientConfig
	nacosDir      string
	logDir        string
	cacheDir      string
)

// 添加全局变量存储服务实例信息，用于心跳检查和重新注册
var (
	registeredServices = make(map[string]*RegisteredServiceInfo)
	serviceCheckMutex  = &sync.Mutex{}
)

// RegisteredServiceInfo 存储已注册服务的信息
type RegisteredServiceInfo struct {
	ServiceName   string
	IP            string
	Port          uint64
	Group         string
	NamingClient  *naming_client.INamingClient
	LastHeartbeat time.Time
}
