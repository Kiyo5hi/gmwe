package hitokoto

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
		{"unknown member", `{"Content":"no","UserID":999}`, "application/json", true, 400},
		{"invalid member", `{"Content":"no","UserID":-1}`, "application/json", true, 400},
		{"forged submitter", `{"Content":"no","SubmittedByUserID":999}`, "application/json", true, 400},
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
	if rows[0].SubmittedByUserID == nil || *rows[0].SubmittedByUserID != int(u.ID) {
		t.Fatal("submitter missing")
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
		{"?user_id=-1", 400, 0, 0}, {"?user_id=abc", 400, 0, 0}, {"?user_id=999", 200, 0, 0},
		{"?from=invalid", 400, 0, 0}, {"?to=2026-02-30", 400, 0, 0}, {"?from=2026-02-02&to=2026-02-01", 400, 0, 0},
		{"?sort=id", 400, 0, 0},
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
	other := auth.User{Username: "other", Name: "Other"}
	if err := engine.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(fmt.Sprintf(`{"Content":"Delegated entry","UserID":%d}`, other.ID)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Writer", "yes")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	var created struct{ Data Hitokoto }
	if res.Code != 201 || json.Unmarshal(res.Body.Bytes(), &created) != nil {
		t.Fatal("delegated write failed")
	}
	if created.Data.UserID != int(other.ID) || created.Data.User.Name != "Other" || created.Data.SubmittedByUserID == nil || *created.Data.SubmittedByUserID != int(u.ID) {
		t.Fatal("attribution mismatch")
	}
	if err := engine.Model(&created.Data).Update("created_at", time.Date(2000, 1, 2, 23, 59, 59, 0, time.UTC)).Error; err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		query string
		total int
	}{
		{fmt.Sprintf("?user_id=%d&from=2000-01-02&to=2000-01-02&q=delegated", other.ID), 1},
		{"?to=2000-01-01", 0}, {"?from=2000-01-03", 23}, {"?sort=oldest", 24},
	} {
		res := httptest.NewRecorder()
		r.ServeHTTP(res, httptest.NewRequest("GET", "/list"+tc.query, nil))
		var got struct {
			Data  []Hitokoto
			Total int
		}
		if res.Code != 200 || json.Unmarshal(res.Body.Bytes(), &got) != nil || got.Total != tc.total {
			t.Fatalf("filter failed %s: %s", tc.query, res.Body.String())
		}
		if tc.query == "?sort=oldest" && got.Data[0].ID != rows[0].ID {
			t.Fatal("oldest order failed")
		}
	}
	if err := engine.Delete(&other).Error; err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest("POST", "/", strings.NewReader(fmt.Sprintf(`{"Content":"Deleted member","UserID":%d}`, other.ID)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Writer", "yes")
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != 400 {
		t.Fatal("deleted member accepted")
	}
}
