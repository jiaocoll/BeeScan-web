package middleware

import (
	"Beescan/controller"
	"Beescan/core/config"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
创建人员：云深不知处
创建时间：2022/4/1
程序功能：
*/
var cookieValue = base64.StdEncoding.EncodeToString([]byte(config.GlobalConfig.UserPassConfig.UserName + "-" + config.GlobalConfig.UserPassConfig.PassWord))

func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/info" || c.Request.URL.Path == "/scan" || c.Request.URL.Path == "/assets" || c.Request.URL.Path == "/csv" || c.Request.URL.Path == "/task" || c.Request.URL.Path == "/ipdetail" || c.Request.URL.Path == "/vul" || c.Request.URL.Path == "/vuldetail" || c.Request.URL.Path == "/logs" || c.Request.URL.Path == "/nodelog" {
			// 获取客户端cookie并校验
			if cookie, err := c.Cookie("Beescan"); err == nil {
				if cookie == controller.CookieValue {
					c.Next()
					return
				}
			}
			// 返回错误
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录！"})
			// 若验证不通过，不再调用后续的函数处理
			c.Abort()
			return
		}
		c.Next()
		return
	}
}
