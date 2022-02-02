package main

import (
	"Beescan/routers"
	"github.com/gin-gonic/gin"
)

/*
创建人员：云深不知处
创建时间：2022/1/7
程序功能：主程序
*/
func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	r := routers.SetupRouter()
	r.Run(":9090")
}
