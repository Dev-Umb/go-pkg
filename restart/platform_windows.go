//go:build windows
// +build windows

package restart

import (
	"os/exec"
	"syscall"
)

// setPlatformSpecificAttributes 设置Windows平台特定的进程属性
func setPlatformSpecificAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
