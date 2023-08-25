package ping

import (
	"net/http"

	"acey.k1yoshi.com/api/utils/requests"
	"github.com/gin-gonic/gin"
)

type PingApi struct{}

func (PingApi) All(c *gin.Context) {
	c.JSON(http.StatusOK, requests.Response("pong", requests.PageMetadata{}))
}
