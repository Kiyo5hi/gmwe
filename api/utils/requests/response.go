package requests

import (
	"github.com/gin-gonic/gin"
)

func Response(d any, pm PageMetadata) (h gin.H) {
	return gin.H{
		"Data":         d,
		"PageMetadata": pm,
	}
}

func Error(em string, d error) (h gin.H) {
	return gin.H{
		"Message": em,
		"Details": d.Error(),
	}
}
