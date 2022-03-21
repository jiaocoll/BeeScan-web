package handler

import (
	"Beescan/msg"
	"Beescan/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/2/23
程序功能：中间件
*/

// middleware
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		var data interface{}
		var errStr string

		code = msg.SuccessCode
		token := c.Request.Header.Get("Authorization")

		if token == "" {
			// 非登录状态
			code = msg.ErrCode
			errStr = "请登录后操作"
		} else {
			claims, err := utils.ParseToken(token)
			if err != nil {
				//	token 校验不通过
				code = msg.ErrCode
				errStr = "身份验证失败，请重新登录"
			} else if time.Now().Unix() > claims.ExpiresAt {
				//	token 已过期
				code = msg.ErrCode
				errStr = "身份信息已过期，请重新登录"
			}
		}

		if code != msg.SuccessCode {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": code,
				"msg":  errStr,
				"data": data,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
