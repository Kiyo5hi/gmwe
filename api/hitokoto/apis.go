package hitokoto

import (
	"encoding/json"
	"gmwe/api/auth"
	"io"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"

	"gmwe/api/utils/requests"

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
	writer, ok := c.Get("writer")
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	media, _, errMedia := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if errMedia != nil || media != "application/json" {
		c.AbortWithStatus(415)
		return
	}
	var input struct {
		Content string
		UserID  *int
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF {
		c.AbortWithStatus(400)
		return
	}
	id := writer.(auth.Writer).UserID
	if input.UserID != nil && *input.UserID != id {
		c.AbortWithStatus(403)
		return
	}
	content := strings.TrimSpace(input.Content)
	if content == "" || !utf8.ValidString(content) || utf8.RuneCountInString(content) > 2000 {
		c.AbortWithStatus(400)
		return
	}
	h := Hitokoto{Content: content, UserID: id}
	err := hitokotoService.CreateHitokoto(&h)
	if err != nil {
		c.AbortWithStatusJSON(err.Status(), requests.Error("Could not create hitokoto", err))
		return
	}
	c.JSON(http.StatusCreated, requests.Response(h, requests.PageMetadata{}))
}
