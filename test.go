package main

import (
	"fmt"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/1/1
程序功能：测试单元
*/

func main() {
	//var ports []string
	//var aliveips []string
	//tmp1 := "80"
	//tmp2 := "buaa.edu.cn"
	//ports = strings.Split(tmp1, ",")
	//aliveips = strings.Split(tmp2, ",")
	//
	//var targets string
	//targets = "test1,"
	////每一个ip和端口构成一个扫描目标,组成目标集合
	//for _, p := range ports {
	//	for _, i := range aliveips {
	//		targets += fmt.Sprintf("%s:%s,", i, p)
	//	}
	//}
	//
	//fmt.Println(targets)
	////将目标送进redis消息队列中
	//conn, err := db.RedisInit()
	//if err != nil {
	//	log.Println(err)
	//}
	//jsjob, err := json.MarshalBinary(targets)
	//if err != nil {
	//	log.Println(err)
	//}
	////fmt.Println(jsjob)
	//_ = db.AddJob(conn, jsjob, "BeeScanQueue_node_1")

	//teststr := `domain="BeeScan.com" && title="BeeScan"`
	//a := strings.Split(teststr, " && ")
	//for _, v := range a {
	//	tmpv1 := strings.Split(v, "=\"")
	//	key := tmpv1[0]
	//	//key = strings.Replace(key, " ", "", -1)
	//	tmpv2 := strings.Split(tmpv1[1], "\"")
	//	value := strings.Replace(tmpv2[0], "\"", "", -1)
	//	fmt.Println(key, value)
	//}

	current := time.Now().Unix()

	fmt.Println(current)

	BeforeDate := "2022-01-21 12:48:00"
	loc, _ := time.LoadLocation("Local") //获取时区
	tmp, _ := time.ParseInLocation("2006-01-02 15:04:05", BeforeDate, loc)
	timestamp := tmp.Unix() //转化为时间戳 类型是int64

	fmt.Println(timestamp)

	res := (current - timestamp) / 86400 //相差值

	fmt.Println(res)

}
