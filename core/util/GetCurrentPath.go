package util

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

/*
创建人员：云深不知处
创建时间：2022/3/29
程序功能：获取当前目录位置
*/

// GetCurrentDirectory 获取当前目录位置
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Println(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}
