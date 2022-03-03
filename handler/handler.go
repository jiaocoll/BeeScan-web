package handler

import (
	"Beescan/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"log"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/2/23
程序功能：中间件
*/

type UserClaims struct {
	Username string `json:"username"`
	Password string `json:"password"`
	jwt.StandardClaims
}

var (
	//token秘钥
	secret = []byte("Beescan")
	//该路由下不检验token
	noVerify = []interface{}{"/", "/login"}
	//token有效时间
	effectTime = 2 * time.Hour
)

// GenerateToken 生成Token
func GenerateToken(claims *UserClaims) string {
	claims.ExpiresAt = time.Now().Add(effectTime).Unix()

	//生成token
	sign, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		log.Println(err)
	}
	return sign
}

// JwtVerify 验证Token
func JwtVerify(c *gin.Context) {
	//过滤是否验证token
	if utils.IsContailArr(noVerify, c.Request.RequestURI) {
		return
	}
	token := c.GetHeader("token")
	if token == "" {
		panic("token is not exist!")
	}
	c.Set("user", ParseToken(token))
}

// ParseToken 解析Token
func ParseToken(tokenString string) *UserClaims {
	//解析token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) { return secret, nil })
	if err != nil {
		log.Println(err)
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		log.Println("token is valid")
	}
	return claims
}
