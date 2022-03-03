package main

import (
	"Beescan/core/runner"
	"fmt"
)

/*
创建人员：云深不知处
创建时间：2022/1/1
程序功能：测试单元
*/

func main() {
	targets := []string{"http://www.baidu.com"}
	fmt.Println("111111111111111")
	r := runner.NewRunner(targets)
	fmt.Println("222222222222222")
	r.ParsePocs()
	fmt.Println("3333333333333333")
	r.RunPoc()
	fmt.Println("4444444444444444")
}
