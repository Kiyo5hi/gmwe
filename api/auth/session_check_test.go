package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSessionCheckDoesNotHydrateProfile(t *testing.T) {
	f := newFixture(t)
	*f.userinfo = func(w http.ResponseWriter, r *http.Request) {
		t.Error("session check must not fetch UserInfo")
		w.WriteHeader(503)
	}
	r := gin.New()
	f.a.Routes(r)
	r.GET("/seed", func(c *gin.Context) {
		f.a.Sessions.Put(c.Request.Context(), "access", f.token(t, "at+jwt", nil))
		c.Status(204)
	})
	h := f.a.Middleware(r)
	seed := httptest.NewRecorder()
	h.ServeHTTP(seed, httptest.NewRequest("GET", "https://gmwe.test/seed", nil))
	for _, signedIn := range []bool{false, true} {
		req := httptest.NewRequest("GET", "https://gmwe.test/api/auth/session", nil)
		want := 401
		if signedIn {
			req.AddCookie(seed.Result().Cookies()[0])
			want = 204
		}
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if res.Code != want || res.Body.Len() != 0 {
			t.Fatalf("session response: %d, body bytes: %d", res.Code, res.Body.Len())
		}
		if res.Header().Get("Cache-Control") != "no-store" && res.Header().Get("Cache-Control") != "private, no-store" {
			t.Fatal("session check must not be cached")
		}
	}
}
