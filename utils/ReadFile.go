package utils

import (
	"Beescan/core/db"
	"Beescan/core/scan"
	"bufio"
	"github.com/olivere/elastic/v7"
	"io"
	"log"
	"os"
	"strings"
)

/*
创建人员：云深不知处
创建时间：2022/3/29
程序功能：读取文件并执行绑定函数
*/

func ReadLine(es *elastic.Client, fileName string, handler func(string) scan.NucleiOutput, taskname string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		output := handler(line)
		if output.Template != "" {
			output.TaskName = taskname
			output.ID = output.TaskName + "-" + output.Host + "-" + output.IP + output.Info.Name
			db.EsAdd(es, output)
		}
		if err != nil {
			if err == io.EOF {
				err1 := f.Close()
				if err1 != nil {
					log.Println("Close file:", err1)
				}
				return nil
			}
			return err
		}
	}
}
