// Package logger **/
/**
* @Email: i@umb.ink
 */
package logger

import (
	"os"
	"runtime"
)

var (
	goVersion string
)
var (
	hostName        string
	buildAppVersion string
	buildUser       string
	buildHost       string
	buildTime       string
)

func init() {
	name, err := os.Hostname()
	if err != nil {
		name = "unknown"
	}
	hostName = name
	goVersion = runtime.Version()
}

// AppVersion get buildAppVersion
func AppVersion() string {
	return buildAppVersion
}

// BuildTime get buildTime
func BuildTime() string {
	return buildTime
}

// BuildUser get buildUser
func BuildUser() string {
	return buildUser
}

// BuildHost get buildHost
func BuildHost() string {
	return buildHost
}

// HostName get host name
func HostName() string {
	return hostName
}

// GoVersion get go version
func GoVersion() string {
	return goVersion
}
