package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestExistingSessionProfile(t *testing.T) {
	for _, tc := range []struct {
		name, body, state string
		status            int
		loaded            bool
	}{
		{"profile", `{"sub":"allowed","given_name":"First","family_name":"Last"}`, "ready", 200, true},
		{"empty names", `{"sub":"allowed","given_name":"","family_name":""}`, "ready", 200, true},
		{"old grant", `{"sub":"allowed"}`, "consent_required", 200, false},
		{"different user", `{"sub":"someone-else","given_name":"Wrong"}`, "unavailable", 200, false},
		{"provider failure", `{}`, "unavailable", 503, false},
		{"malformed", `{"sub":"allowed","given_name":123}`, "unavailable", 200, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			access := f.token(t, "at+jwt", nil)
			calls := 0
			*f.userinfo = func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Header.Get("Authorization") != "Bearer "+access {
					t.Error("missing provider credential")
				}
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}
			r := gin.New()
			f.a.Routes(r)
			r.GET("/seed", func(c *gin.Context) {
				f.a.Sessions.Put(c.Request.Context(), "access", access)
				// The preceding version incorrectly marked an absent profile loaded.
				f.a.Sessions.Put(c.Request.Context(), "profile_loaded", true)
				c.Status(204)
			})
			h := f.a.Middleware(r)
			seed := httptest.NewRecorder()
			h.ServeHTTP(seed, httptest.NewRequest("GET", "https://gmwe.test/seed", nil))
			cookie := seed.Result().Cookies()[0]
			for range 2 {
				req := httptest.NewRequest("GET", "https://gmwe.test/api/auth/me", nil)
				req.AddCookie(cookie)
				res := httptest.NewRecorder()
				h.ServeHTTP(res, req)
				var result struct{ Profile accountProfile }
				if res.Code != 200 || json.Unmarshal(res.Body.Bytes(), &result) != nil || result.Profile.State != tc.state || result.Profile.Loaded != tc.loaded {
					t.Fatalf("bad profile status: %s", res.Body.String())
				}
				if tc.name == "profile" && (result.Profile.FirstName != "First" || result.Profile.LastName != "Last") {
					t.Fatal("names not hydrated")
				}
				if !tc.loaded && result.Profile.FirstName != "" {
					t.Fatal("unverified name accepted")
				}
				if cookies := res.Result().Cookies(); len(cookies) > 0 {
					cookie = cookies[0]
				}
			}
			if calls != 1 {
				t.Fatal("userinfo retry/cache not bounded")
			}
		})
	}
}
