package auth

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"gmwe/api/consts"
	"gmwe/api/db"
)

func TestMembersRequireRealSessionMiddleware(t *testing.T) {
	consts.DB_URI = "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "members.db")) + "?mode=rwc"
	engine := db.DB().Engine
	pool, _ := engine.DB()
	t.Cleanup(func() { pool.Close() })
	if err := engine.AutoMigrate(&User{}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Create(&User{Username: "member", Name: "Member", Password: "must-not-leak"}).Error; err != nil {
		t.Fatal(err)
	}
	f := newFixture(t)
	r := gin.New()
	r.GET("/api/v1/users", f.a.RequireWriter(), (UsersAPI{}).Get)
	r.GET("/seed", func(c *gin.Context) {
		f.a.Sessions.Put(c.Request.Context(), "access", f.token(t, "at+jwt", nil))
		c.Status(204)
	})
	h := f.a.Middleware(r)
	seed := httptest.NewRecorder()
	h.ServeHTTP(seed, httptest.NewRequest("GET", "https://gmwe.test/seed", nil))
	for _, signedIn := range []bool{false, true} {
		req := httptest.NewRequest("GET", "https://gmwe.test/api/v1/users", nil)
		if signedIn {
			req.AddCookie(seed.Result().Cookies()[0])
		}
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if !signedIn {
			if res.Code != 401 {
				t.Fatal("anonymous members exposed")
			}
			continue
		}
		var body struct{ Data []map[string]any }
		if res.Code != 200 || json.Unmarshal(res.Body.Bytes(), &body) != nil || len(body.Data) != 1 {
			t.Fatalf("signed-in members failed: %d", res.Code)
		}
		if body.Data[0]["Name"] != "Member" || body.Data[0]["Password"] != nil {
			t.Fatal("unsafe or incorrect member response")
		}
	}
}
