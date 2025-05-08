package nacos_sdk

var (
	nacosAddress   string
	nacosPort      uint64
	nacosNameSpace string
	projectName    string
	nacosGroup     string
	rpcPort        string
)

type NacosConfig struct {
	NacosAddress   string
	NacosPort      uint64
	NacosNameSpace string
	ProjectName    string
	NacosGroup     string
	RpcPort        string
}

func InitNacosSDK(config NacosConfig) {
	nacosAddress = config.NacosAddress
	nacosPort = config.NacosPort
	nacosNameSpace = config.NacosNameSpace
	projectName = config.ProjectName
	nacosGroup = config.NacosGroup
	rpcPort = config.RpcPort
}
