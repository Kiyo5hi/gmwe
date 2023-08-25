package auth

import (
	"net/http"

	"gmwe/api/utils/requests"

	"github.com/gin-gonic/gin"
)

type UsersAPI struct{}

var userService = new(UserService)

func (UsersAPI) Get(c *gin.Context) {
	us, err := userService.GetUsers()
	if err != nil {
		c.AbortWithStatusJSON(err.Status(), requests.Error("Could not get users", err))
		return
	}
	c.JSON(http.StatusOK, requests.Response(us, requests.PageMetadata{}))
}
