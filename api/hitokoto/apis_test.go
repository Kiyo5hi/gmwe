package hitokoto

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gmwe/api/auth"
	"gmwe/api/consts"
	"gmwe/api/db"
)

func TestWriteAuthorAndInputBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	consts.DB_URI = "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "test.db")) + "?mode=rwc"
	engine := db.DB().Engine
	pool, _ := engine.DB()
	t.Cleanup(func() { pool.Close() })
	if err := engine.AutoMigrate(&auth.User{}, &Hitokoto{}); err != nil {
		t.Fatal(err)
	}
	u := auth.User{Username: "writer", Name: "Writer"}
	if err := engine.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	api := HitokotoAPI{}
	r.POST("/", func(c *gin.Context) {
		if c.GetHeader("X-Test-Writer") == "yes" {
			c.Set("writer", auth.Writer{UserID: int(u.ID)})
		}
	}, api.Post)
	for _, tc := range []struct {
		name, body, media string
		writer            bool
		want              int
	}{
		{"anonymous", `{"Content":"no"}`, "application/json", false, 401},
		{"forged author", `{"Content":"no","UserID":999}`, "application/json", true, 403},
		{"mass assignment", `{"Content":"no","ID":999}`, "application/json", true, 400},
		{"trailing document", `{"Content":"no"} {}`, "application/json", true, 400},
		{"blank", `{"Content":"  "}`, "application/json", true, 400},
		{"too long", `{"Content":"` + strings.Repeat("a", 2001) + `"}`, "application/json", true, 400},
		{"body cap", `{"Content":"` + strings.Repeat("a", 9000) + `"}`, "application/json", true, 400},
		{"wrong media", `{"Content":"no"}`, "text/plain", true, 415},
		{"valid", `{"Content":" Valid entry "}`, "application/json", true, 201},
		{"duplicate", `{"Content":"Valid entry"}`, "application/json", true, 409},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.media)
			if tc.writer {
				req.Header.Set("X-Test-Writer", "yes")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("got %d want %d", w.Code, tc.want)
			}
		})
	}
	var rows []Hitokoto
	if err := engine.Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].UserID != int(u.ID) || rows[0].Content != "Valid entry" {
		t.Fatal("write boundary or author mapping failed")
	}
}
