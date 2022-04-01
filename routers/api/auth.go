package api

import (
	"Beescan/models"
	"Beescan/msg"
	"Beescan/utils"
	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

/*
创建人员：云深不知处
创建时间：2022/3/29
程序功能：
*/

/*
 获得一个Token
*/
func GetAuth(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	valid := validation.Validation{}
	a := models.Auth{Username: username, Password: password}
	ok, _ := valid.Valid(&a)

	data := make(map[string]interface{})
	code := msg.INVALID_PARAMS
	if ok {
		isExist := true //corll.GetUserByNameAndPassword(username, password) //models.CheckAuth(username, password)
		if isExist {
			token, err := utils.GenerateToken(username, password)
			if err != nil {
				code = msg.ERROR_AUTH_TOKEN
			} else {
				data["token"] = token

				code = msg.SUCCESS
			}

		} else {
			code = msg.ERROR_AUTH
		}
	} else {
		for _, err := range valid.Errors {
			log.Println(err.Key, err.Message)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  msg.GetMsg(code),
		"data": data,
	})
}
