package nacos_sdk

var (
	NacosAddress   string
	NacosPort      uint64
	NacosNameSpace string
)

func InitNacosSDK(nacosAddress string, nacosPort uint64, nacosNameSpace string) {
	NacosAddress = nacosAddress
	NacosPort = nacosPort
	nacosNameSpace = nacosNameSpace
}
