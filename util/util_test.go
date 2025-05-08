package util

import (
	"testing"
)

const path = "/tmp/test"

func TestCreateDir(t *testing.T) {
	err := CreateDir(path)
	if err == nil {
		t.Logf("目录已存在")
	} else {
		t.Logf("目录创建成功")
	}
}

func TestDirExist(t *testing.T) {

	ok, _ := DirExist(path)
	if ok {
		t.Logf("目录存在")
	} else {
		t.Logf("目录不存在")
	}
}
