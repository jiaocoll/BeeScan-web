package main

import (
	"Beescan/core/util"
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

/*
创建人员：云深不知处
创建时间：2022/1/1
程序功能：测试单元
*/

func main() {
	path := util.GetCurrentDirectory()
	fmt.Println(path)
}

func ReadLine(fileName string, handler func(string)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		handler(line)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

func Print(line string) {
	fmt.Println(line)
}
