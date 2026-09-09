package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestPersistentSessionRefreshAndLogout(t *testing.T) {
	f := newFixture(t)
	path := filepath.Join(t.TempDir(), "auth", "sessions.db")
	closeStore, err := f.a.PersistSessions(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeStore != nil {
			closeStore()
		}
	})
	expired := f.token(t, "at+jwt", map[string]any{"exp": time.Now().Add(-time.Minute).Unix()})
	fresh := f.token(t, "at+jwt", nil)
	var calls atomic.Int32
	*f.exchange = func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		req.ParseForm()
		if req.Form.Get("grant_type") != "refresh_token" || req.Form.Get("refresh_token") != "fixture-refresh" || req.Form.Get("client_id") != "gmwe" {
			t.Error("unexpected refresh exchange")
		}
		json.NewEncoder(w).Encode(map[string]any{"access_token": fresh, "refresh_token": "fixture-rotated", "token_type": "Bearer", "expires_in": 3600})
	}
	r := gin.New()
	r.GET("/seed", func(c *gin.Context) {
		f.a.Sessions.Put(c.Request.Context(), "access", expired)
		f.a.Sessions.Put(c.Request.Context(), "refresh", "fixture-refresh")
		f.a.Sessions.Put(c.Request.Context(), "subject", "allowed")
		f.a.Sessions.Put(c.Request.Context(), "csrf", "fixture-csrf")
		c.Status(204)
	})
	f.a.Routes(r)
	h := f.a.Middleware(r)
	seed := httptest.NewRecorder()
	h.ServeHTTP(seed, httptest.NewRequest("GET", "https://gmwe.test/seed", nil))
	cookie := seed.Result().Cookies()[0]
	if cookie.MaxAge <= 0 || cookie.Expires.Before(time.Now().Add(6*24*time.Hour)) || !cookie.HttpOnly || !cookie.Secure {
		t.Fatal("persistent secure cookie missing")
	}
	closeStore()
	// Reopening the store simulates restart without importing any session state.
	closeStore, err = f.a.PersistSessions(path)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			req := httptest.NewRequest("GET", "https://gmwe.test/api/auth/me", nil)
			req.AddCookie(cookie)
			res := httptest.NewRecorder()
			h.ServeHTTP(res, req)
			if res.Code != 200 || strings.Contains(res.Body.String(), "fixture-rotated") || strings.Contains(res.Body.String(), fresh) {
				t.Error("restored session failed or leaked credentials")
			}
		})
	}
	group.Wait()
	if calls.Load() != 1 {
		t.Fatal("concurrent renewal rotated a token more than once")
	}
	req := httptest.NewRequest("POST", "https://gmwe.test/api/auth/logout", nil)
	req.AddCookie(cookie)
	req.Header.Set("Origin", "https://gmwe.test")
	req.Header.Set("X-CSRF-Token", "fixture-csrf")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != 204 {
		t.Fatalf("logout: %d", res.Code)
	}
	req = httptest.NewRequest("GET", "https://gmwe.test/api/auth/me", nil)
	req.AddCookie(cookie)
	res = httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != 401 {
		t.Fatal("logged-out cookie can be replayed")
	}
}

func TestRenewalFailureBoundaries(t *testing.T) {
	for _, mode := range []string{"outage", "revoked", "access-denied", "wrong-subject"} {
		t.Run(mode, func(t *testing.T) {
			f := newFixture(t)
			expired := f.token(t, "at+jwt", map[string]any{"exp": time.Now().Add(-time.Minute).Unix()})
			*f.exchange = func(w http.ResponseWriter, req *http.Request) {
				switch mode {
				case "outage":
					w.WriteHeader(503)
				case "revoked":
					w.WriteHeader(400)
					w.Write([]byte(`{"error":"invalid_grant"}`))
				case "access-denied":
					w.WriteHeader(403)
					w.Write([]byte(`{"error":"access_denied"}`))
				default:
					json.NewEncoder(w).Encode(map[string]any{"access_token": f.token(t, "at+jwt", map[string]any{"sub": "other"}), "refresh_token": "rotated", "token_type": "Bearer"})
				}
			}
			r := gin.New()
			r.GET("/seed", func(c *gin.Context) {
				f.a.Sessions.Put(c.Request.Context(), "access", expired)
				f.a.Sessions.Put(c.Request.Context(), "refresh", "fixture-refresh")
				f.a.Sessions.Put(c.Request.Context(), "subject", "allowed")
				c.Status(204)
			})
			f.a.Routes(r)
			h := f.a.Middleware(r)
			seed := httptest.NewRecorder()
			h.ServeHTTP(seed, httptest.NewRequest("GET", "https://gmwe.test/seed", nil))
			req := httptest.NewRequest("GET", "https://gmwe.test/api/auth/me", nil)
			req.AddCookie(seed.Result().Cookies()[0])
			res := httptest.NewRecorder()
			h.ServeHTTP(res, req)
			want := 401
			if mode == "outage" {
				want = 503
			}
			if res.Code != want {
				t.Fatalf("got %d want %d", res.Code, want)
			}
		})
	}
}
