package utils

import (
	"Beescan/core/db"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/2/23
程序功能：工具包
*/

type Asset struct {
	URL   string
	IP    string
	Port  string
	Title string
}

func GetdataTocsv(Datas []db.Output) (string, error) {
	var res []*Asset
	for _, v := range Datas {
		res = append(res, &Asset{
			URL:   v.Domain,
			IP:    v.Ip,
			Port:  v.Port,
			Title: v.Webbanner.Title,
		})
	}
	strTime := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("资产%s.csv", strTime)
	xlsFile, err := os.OpenFile("./"+filename, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		log.Println(err)
	}
	defer xlsFile.Close()

	xlsFile.WriteString("\xEF\xBB\xBF")
	wstr := csv.NewWriter(xlsFile)
	wstr.Write([]string{"URL", "IP", "Port", "Title"})

	for _, data := range res {
		wstr.Write([]string{data.URL, data.IP, data.Port, data.Title})
	}
	wstr.Flush()
	return filename, nil
}

func IsContailArr(substr []interface{}, target string) bool {
	for _, v := range substr {
		if target == v.(string) {
			return true
		}
	}
	return false
}
