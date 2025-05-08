// Package restart 提供服务重启功能
package restart

import (
	"github.com/Dev-Umb/go-pkg/logger"
	"os"
	"os/exec"
	"time"
)

var (
	// 是否已经触发重启
	restarting = false
	// 重启延迟时间（秒）
	restartDelay = 3
	// 服务启动时间
	startTime = time.Now()
	// 服务稳定运行时间（秒），只有超过这个时间才允许重启
	stableRunningTime = 60
)

// RestartService 重启当前服务
// 通过启动一个新的进程并退出当前进程来实现重启
func RestartService() {
	// 如果已经在重启过程中，则不再重复触发
	if restarting {
		return
	}

	// 检查服务是否刚刚启动
	runningDuration := time.Since(startTime).Seconds()
	if runningDuration < float64(stableRunningTime) {
		logger.Warnf("服务刚启动 %.2f 秒，小于稳定运行时间 %d 秒，忽略此次重启请求", runningDuration, stableRunningTime)
		return
	}

	restarting = true
	logger.Info("配置变更，服务将在 3 秒后重启...")

	// 延迟几秒再重启，确保所有日志都已写入
	go func() {
		time.Sleep(time.Duration(restartDelay) * time.Second)

		// 获取当前可执行文件路径
		execPath, err := os.Executable()
		if err != nil {
			logger.Errorf("获取可执行文件路径失败: %v", err)
			return
		}

		// 获取当前工作目录
		workDir, err := os.Getwd()
		if err != nil {
			logger.Errorf("获取工作目录失败: %v", err)
			return
		}

		// 获取命令行参数
		args := os.Args[1:]

		// 创建新进程
		cmd := exec.Command(execPath, args...)
		cmd.Dir = workDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		// 根据不同操作系统设置进程属性
		setPlatformSpecificAttributes(cmd)

		// 启动新进程
		err = cmd.Start()
		if err != nil {
			logger.Errorf("启动新进程失败: %v", err)
			return
		}

		logger.Infof("新进程已启动，PID: %d", cmd.Process.Pid)

		// 退出当前进程
		os.Exit(0)
	}()
}

// SetRestartDelay 设置重启延迟时间（秒）
func SetRestartDelay(seconds int) {
	if seconds > 0 {
		restartDelay = seconds
	}
}

// SetStableRunningTime 设置服务稳定运行时间（秒）
func SetStableRunningTime(seconds int) {
	if seconds > 0 {
		stableRunningTime = seconds
	}
}
