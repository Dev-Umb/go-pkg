package util

import (
	"context"
	"net"

	"github.com/Dev-Umb/go-pkg/logger"
)

// 获取本机IP地址
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Errorf(context.Background(), "net.InterfaceAddrs: %v", err)
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
