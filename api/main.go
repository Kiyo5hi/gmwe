package main

import (
	"fmt"

	"acey.k1yoshi.com/api/auth"
	"acey.k1yoshi.com/api/consts"
	"acey.k1yoshi.com/api/hitokoto"
	"acey.k1yoshi.com/api/ping"
	"acey.k1yoshi.com/api/utils/db"
	"acey.k1yoshi.com/api/utils/middlewares"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(consts.ENV)
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// Applying middlewares
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.RequestIDMiddleware())

	db.InitDB()
	v1 := r.Group("/api/v1")
	{
		hitokotoAPI := new(hitokoto.HitokotoAPI)
		v1.GET("/hitokoto", hitokotoAPI.Get)
		v1.POST("/hitokoto", hitokotoAPI.Post)

		usersAPI := new(auth.UsersAPI)
		v1.GET("/users", usersAPI.Get)

		pingAPI := new(ping.PingApi)
		v1.Any("/ping", pingAPI.All)
	}

	r.Run(fmt.Sprintf(":%s", consts.PORT))
}
