package auth

import (
	"crypto/md5"
	"encoding/hex"
	"opsServer/service"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var expTime = 30 * time.Minute

var key = ""

var authStr = ""

var Encryption = md5.Sum([]byte(key + "_" + authStr))

var apiAuth = hex.EncodeToString(Encryption[:])

var jwtKey = []byte(key + "_" + authStr)

type AuthData struct {
	SID  string `json:"secret_id"`
	SKey string `json:"secret_key"`
}

var secretID = ""

var secretKey = ""

type Claims struct {
	SID string `json:"secret_id"`
	jwt.StandardClaims
}

func GetToken(c *gin.Context) {
	var user AuthData
	if err := c.ShouldBindJSON(&user); err != nil {
		service.Fail[string](c, service.Unauthorized)
		return
	}
	if user.SID == secretID && user.SKey == secretKey {
		expirationTime := time.Now().Add(expTime)
		claims := &Claims{
			SID: user.SID,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			service.Fail[string](c, service.Unauthorized)
			return
		}
		service.OK(c, tokenString)
	} else {
		service.Fail[string](c, service.Unauthorized)
	}
}

func AUTH() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("API-AUTH")
		if tokenString == "" {
			service.Fail[string](c, service.Unauthorized)
			return
		}
		if tokenString == apiAuth {
			c.Next()
			return
		}
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			service.Fail[string](c, service.Unauthorized)
			c.Abort()
			return
		}

		if !token.Valid {
			service.Fail[string](c, service.Unauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}
