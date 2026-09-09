package hitokoto

import (
	"encoding/json"
	"fmt"
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
	for i := range 22 {
		if err := engine.Create(&Hitokoto{Content: fmt.Sprintf("list fixture %d", i), UserID: int(u.ID)}).Error; err != nil {
			t.Fatal(err)
		}
	}
	r.GET("/list", api.List)
	for _, tc := range []struct {
		query                string
		status, count, total int
	}{
		{"", 200, 20, 23}, {"?page=2", 200, 3, 23}, {"?page=3", 200, 0, 23},
		{"?q=VALID", 200, 1, 1}, {"?q=%25", 200, 0, 0}, {"?q=%27", 200, 0, 0},
		{"?page=0", 400, 0, 0}, {"?page=-1", 400, 0, 0}, {"?page=abc", 400, 0, 0}, {"?q=" + strings.Repeat("a", 101), 400, 0, 0},
	} {
		res := httptest.NewRecorder()
		r.ServeHTTP(res, httptest.NewRequest("GET", "/list"+tc.query, nil))
		if res.Code != tc.status {
			t.Fatalf("list status %d", res.Code)
		}
		if tc.status != 200 {
			continue
		}
		var result struct {
			Data  []Hitokoto
			Total int
		}
		if json.Unmarshal(res.Body.Bytes(), &result) != nil || len(result.Data) != tc.count || result.Total != tc.total {
			t.Fatal("list pagination or literal search failed")
		}
		for i, item := range result.Data {
			if item.User.Name != "Writer" || (i > 0 && item.ID >= result.Data[i-1].ID) {
				t.Fatal("list author/order mismatch")
			}
		}
	}
}
