package hitokoto

import (
	"encoding/json"
	"gmwe/api/auth"
	"gmwe/api/db"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"gmwe/api/utils/requests"

	"github.com/gin-gonic/gin"
)

type HitokotoAPI struct{}

var hitokotoService = new(HitokotoService)

func (HitokotoAPI) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	query := strings.TrimSpace(c.Query("q"))
	if err != nil || page < 1 || page > 100000 || !utf8.ValidString(query) || utf8.RuneCountInString(query) > 100 {
		c.AbortWithStatus(400)
		return
	}
	engine := db.DB().Engine.WithContext(c.Request.Context()).Model(&Hitokoto{})
	if query != "" {
		engine = engine.Where("instr(lower(content), lower(?)) > 0", query)
	}
	var total int64
	if engine.Count(&total).Error != nil {
		c.AbortWithStatus(500)
		return
	}
	items := []Hitokoto{}
	if engine.Order("id DESC").Limit(20).Offset((page-1)*20).Preload("User").Find(&items).Error != nil {
		c.AbortWithStatus(500)
		return
	}
	c.JSON(200, gin.H{"Data": items, "Total": total, "Page": page})
}

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
