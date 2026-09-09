package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gmwe/api/auth"
	"gmwe/api/consts"
	"gmwe/api/hitokoto"
	"gmwe/api/ping"
	"gmwe/api/utils/db"
	"gmwe/api/utils/middlewares"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(consts.ENV)
	r := gin.New()
	r.Use(gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) { c.AbortWithStatus(500) }))
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// Applying middlewares
	r.Use(middlewares.RequestIDMiddleware())

	if err := db.InitDB(); err != nil {
		panic("database initialization failed")
	}
	var login *auth.OIDCAuth
	if path := os.Getenv("OIDC_CONFIG_FILE"); path != "" {
		cfg, err := auth.LoadOIDCConfig(path)
		if err != nil {
			panic("invalid OIDC configuration")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		login, err = auth.NewOIDC(ctx, cfg)
		cancel()
		if err != nil {
			panic("OIDC initialization failed")
		}
		login.Routes(r)
	}
	v1 := r.Group("/api/v1")
	{
		hitokotoAPI := new(hitokoto.HitokotoAPI)
		v1.GET("/hitokoto", login.RequireWriter(), hitokotoAPI.Get)
		v1.POST("/hitokoto", login.RequireWriter(), hitokotoAPI.Post)

		usersAPI := new(auth.UsersAPI)
		v1.GET("/users", login.RequireWriter(), usersAPI.Get)

		pingAPI := new(ping.PingApi)
		v1.Any("/ping", pingAPI.All)
	}

	var handler http.Handler = r
	if login != nil {
		handler = login.Sessions.LoadAndSave(r)
	}
	server := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%s", consts.PORT), Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 40 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32768}
	if err := server.ListenAndServe(); err != nil {
		panic("HTTP server stopped")
	}
}
