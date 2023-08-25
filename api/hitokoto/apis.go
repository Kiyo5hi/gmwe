package hitokoto

import (
	"net/http"

	"acey.k1yoshi.com/api/utils/requests"
	"github.com/gin-gonic/gin"
)

type HitokotoAPI struct{}

var hitokotoService = new(HitokotoService)

func (HitokotoAPI) Get(c *gin.Context) {
	h, err := hitokotoService.RandomHitokoto()
	if err != nil {
		c.AbortWithStatusJSON(err.Status(), requests.Error("Could not get hitokoto", err))
		return
	}
	c.JSON(http.StatusOK, requests.Response(h, requests.PageMetadata{}))
}

func (HitokotoAPI) Post(c *gin.Context) {
	var h Hitokoto
	c.BindJSON(&h)
	err := hitokotoService.CreateHitokoto(&h)
	if err != nil {
		c.AbortWithStatusJSON(err.Status(), requests.Error("Could not create hitokoto", err))
		return
	}
	c.JSON(http.StatusCreated, requests.Response(h, requests.PageMetadata{}))
}
