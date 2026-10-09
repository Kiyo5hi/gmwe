package auth

import (
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAvatarSessionAndMemberBoundary(t *testing.T) {
	f := newFixture(t)
	r := gin.New()
	f.a.Routes(r)
	r.GET("/seed", func(c *gin.Context) {
		f.a.Sessions.Put(c.Request.Context(), "access", f.token(t, "at+jwt", nil))
		c.Status(204)
	})
	h := f.a.Middleware(r)
	seed := httptest.NewRecorder()
	h.ServeHTTP(seed, httptest.NewRequest("GET", "https://gmwe.test/seed", nil))
	for _, tc := range []struct {
		id     string
		signed bool
		status int
	}{
		{strconv.Itoa(f.a.Config.Writers[0].UserID), false, 401},
		{strconv.Itoa(f.a.Config.Writers[0].UserID), true, 302},
		{"99999", true, 404}, {"invalid", true, 404},
	} {
		req := httptest.NewRequest("GET", "https://gmwe.test/api/auth/avatar/"+tc.id, nil)
		if tc.signed {
			req.AddCookie(seed.Result().Cookies()[0])
		}
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if res.Code != tc.status {
			t.Fatalf("status %d wanted %d", res.Code, tc.status)
		}
		if tc.status == 302 {
			got, _ := url.Parse(res.Header().Get("Location"))
			issuer, _ := url.Parse(f.a.Config.Issuer)
			if got.Host != issuer.Host || got.Path != "/api/users/"+f.a.Config.Writers[0].Subject+"/profile-picture.png" || got.Query().Get("v") == "" || res.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatal("unsafe avatar destination/cache")
			}
		}
	}
}
