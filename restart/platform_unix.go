//go:build !windows
// +build !windows

package restart

import (
	"os/exec"
)

// setPlatformSpecificAttributes 设置Unix/Linux平台特定的进程属性
func setPlatformSpecificAttributes(cmd *exec.Cmd) {
	// Unix/Linux平台不需要特殊设置
}
