package util

import (
	"Beescan/core/db"
	"Beescan/core/scan"
	"bufio"
	"github.com/olivere/elastic/v7"
	"io"
	"os"
	"strings"
)

/*
创建人员：云深不知处
创建时间：2022/3/29
程序功能：
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
		output.TaskName = taskname
		output.ID = output.TaskName + output.Host + "-" + output.IP + output.Info.Name
		db.EsAdd(es, output)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}
