package ginSession

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

//AuthMiddleware 认证中间件
//从请求session data 中获取isLogin

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sd := c.MustGet("session").(Session)
		log.Printf("get session %v from authMD\n", sd)
		isLogin, err := sd.Get("isLogin")
		log.Println(isLogin, err, err == nil && isLogin.(bool))
		if err == nil && isLogin.(bool) {
			//登录状态
			c.Next()
		} else {
			c.Abort()
			c.Redirect(http.StatusFound, "/login")
		}

	}

}
