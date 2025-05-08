package util

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// DirExist 判断目录是否存在
func DirExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateDir 创建目录
func CreateDir(path string) error {
	dirExist, err := DirExist(path)
	if err != nil {
		return err
	}

	if !dirExist {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return errors.New("Create[" + path + "]directory fail:" + err.Error())
		}
		return nil
	}
	return err
}

// StringNowTime ...
func StringNowTime() string {
	now := time.Now()
	nowTime := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	return nowTime
}
