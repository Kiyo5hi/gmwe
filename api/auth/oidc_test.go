package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jose "github.com/go-jose/go-jose/v4"
)

type fixture struct {
	a        *OIDCAuth
	key      *rsa.PrivateKey
	issuer   string
	exchange *http.HandlerFunc
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	var server *httptest.Server
	var exchange http.HandlerFunc
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]any{"issuer": server.URL, "authorization_endpoint": server.URL + "/authorize", "token_endpoint": server.URL + "/token", "device_authorization_endpoint": server.URL + "/device", "jwks_uri": server.URL + "/jwks", "id_token_signing_alg_values_supported": []string{"RS256"}})
		case "/jwks":
			json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "test", Algorithm: "RS256", Use: "sig"}}})
		case "/token", "/device":
			if exchange == nil {
				http.Error(w, "not configured", 400)
			} else {
				exchange(w, r)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	ctx, cancel := context.WithCancel(context.Background())
	a, err := NewOIDC(ctx, OIDCConfig{Issuer: server.URL, ClientID: "gmwe", Resource: "https://gmwe.test/api", Origin: "https://gmwe.test", Writers: []Writer{{Subject: "allowed", UserID: 2, Name: "writer"}}})
	cancel()
	if err != nil {
		t.Fatal(err)
	}
	return fixture{a: a, key: key, issuer: server.URL, exchange: &exchange}
}

func (f fixture) token(t *testing.T, typ string, changes map[string]any) string {
	t.Helper()
	claims := map[string]any{"iss": f.issuer, "sub": "allowed", "aud": []string{f.a.Config.Resource, f.issuer}, "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "client_id": "gmwe", "scope": WriteScope}
	for k, v := range changes {
		if v == nil {
			delete(claims, k)
		} else {
			claims[k] = v
		}
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: f.key}, (&jose.SignerOptions{}).WithType(jose.ContentType(typ)).WithHeader("kid", "test"))
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(claims)
	signed, err := signer.Sign(data)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := signed.CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestAccessBoundary(t *testing.T) {
	f := newFixture(t)
	tests := []struct {
		name    string
		typ     string
		changes map[string]any
		allowed bool
	}{
		{"valid", "at+jwt", nil, true},
		{"id token", "JWT", nil, false},
		{"wrong issuer", "at+jwt", map[string]any{"iss": "https://other.test"}, false},
		{"client audience", "at+jwt", map[string]any{"aud": "gmwe"}, false},
		{"expired", "at+jwt", map[string]any{"exp": time.Now().Add(-time.Minute).Unix()}, false},
		{"no expiry", "at+jwt", map[string]any{"exp": nil}, false},
		{"no issue time", "at+jwt", map[string]any{"iat": nil}, false},
		{"future issue", "at+jwt", map[string]any{"iat": time.Now().Add(time.Hour).Unix()}, false},
		{"not yet valid", "at+jwt", map[string]any{"nbf": time.Now().Add(time.Hour).Unix()}, false},
		{"other client", "at+jwt", map[string]any{"client_id": "other"}, false},
		{"scope prefix", "at+jwt", map[string]any{"scope": "write:hitokoto:admin"}, false},
		{"no scope", "at+jwt", map[string]any{"scope": nil}, false},
		{"other user", "at+jwt", map[string]any{"sub": "not-allowed"}, false},
		{"machine subject", "at+jwt", map[string]any{"sub": "gmwe"}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, err := f.a.VerifyAccess(context.Background(), f.token(t, tc.typ, tc.changes))
			if (err == nil) != tc.allowed {
				t.Fatalf("allowed=%v err=%v", tc.allowed, err)
			}
			if tc.allowed && w.UserID != 2 {
				t.Fatal("wrong author mapping")
			}
		})
	}
	other, _ := rsa.GenerateKey(rand.Reader, 2048)
	f.key = other
	if _, err := f.a.VerifyAccess(context.Background(), f.token(t, "at+jwt", nil)); err == nil {
		t.Fatal("accepted untrusted signature")
	}
}

func TestCookieAndBearerBoundary(t *testing.T) {
	f := newFixture(t)
	r := gin.New()
	r.GET("/seed", func(c *gin.Context) {
		f.a.Sessions.Put(c.Request.Context(), "access", f.token(t, "at+jwt", nil))
		f.a.Sessions.Put(c.Request.Context(), "csrf", "test-csrf")
		c.Status(204)
	})
	r.POST("/write", f.a.RequireWriter(), func(c *gin.Context) { c.Status(204) })
	r.POST("/session", f.a.RequireWriter(), func(c *gin.Context) { c.Status(204) })
	h := f.a.Sessions.LoadAndSave(r)
	seed := httptest.NewRecorder()
	h.ServeHTTP(seed, httptest.NewRequest("GET", "https://gmwe.test/seed", nil))
	cookies := seed.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].Domain != "" || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatal("unsafe session cookie")
	}
	for _, tc := range []struct {
		name, path, origin, csrf, bearer string
		cookie                           bool
		want                             int
	}{
		{"anonymous", "/write", "", "", "", false, 401},
		{"cookie missing csrf", "/write", "https://gmwe.test", "", "", true, 403},
		{"wrong origin", "/write", "https://evil.test", "test-csrf", "", true, 403},
		{"valid cookie", "/write", "https://gmwe.test", "test-csrf", "", true, 204},
		{"retired native bearer", "/write", "", "", f.token(t, "at+jwt", nil), false, 401},
		{"bearer cannot pair", "/session", "", "", f.token(t, "at+jwt", nil), false, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "https://gmwe.test"+tc.path, nil)
			if tc.cookie {
				req.AddCookie(cookies[0])
			}
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("X-CSRF-Token", tc.csrf)
			if tc.bearer != "" {
				req.Header.Set("Authorization", "Bearer "+tc.bearer)
			}
			res := httptest.NewRecorder()
			h.ServeHTTP(res, req)
			if res.Code != tc.want {
				t.Fatalf("got %d want %d", res.Code, tc.want)
			}
		})
	}
}

func TestLoginPKCEAndCallbackState(t *testing.T) {
	f := newFixture(t)
	r := gin.New()
	f.a.Routes(r)
	h := f.a.Sessions.LoadAndSave(r)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest("GET", "https://gmwe.test/api/auth/login", nil))
	u, err := url.Parse(res.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("scope") != "openid "+WriteScope+" offline_access" {
		t.Fatal("login requests unnecessary permissions")
	}
	if q.Get("code_challenge_method") != "S256" || len(q.Get("code_challenge")) < 40 || q.Get("state") == "" || q.Get("nonce") == "" || q.Get("resource") != f.a.Config.Resource || q.Get("redirect_uri") != "https://gmwe.test/api/auth/callback" {
		t.Fatal("missing login protections")
	}
	if strings.Contains(res.Body.String(), "code_verifier") {
		t.Fatal("exposed PKCE verifier")
	}
	req := httptest.NewRequest("GET", "https://gmwe.test/api/auth/callback?state=wrong&code=fake&iss="+url.QueryEscape(f.issuer), nil)
	req.AddCookie(res.Result().Cookies()[0])
	denied := httptest.NewRecorder()
	h.ServeHTTP(denied, req)
	if denied.Code != 400 {
		t.Fatalf("invalid state accepted: %d", denied.Code)
	}
}

func TestDisabledAuthFailsClosed(t *testing.T) {
	var a *OIDCAuth
	r := gin.New()
	r.POST("/", a.RequireWriter(), func(c *gin.Context) { c.Status(204) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/", nil))
	if w.Code != 401 {
		t.Fatal("disabled auth allowed write")
	}
}

func TestSuccessfulCallbackAndReplay(t *testing.T) {
	f := newFixture(t)
	r := gin.New()
	f.a.Routes(r)
	h := f.a.Sessions.LoadAndSave(r)
	start := httptest.NewRecorder()
	h.ServeHTTP(start, httptest.NewRequest("GET", "https://gmwe.test/api/auth/login", nil))
	u, _ := url.Parse(start.Header().Get("Location"))
	q := u.Query()
	access := f.token(t, "at+jwt", nil)
	hash := sha256.Sum256([]byte(access))
	id := f.token(t, "JWT", map[string]any{"aud": "gmwe", "nonce": q.Get("nonce"), "at_hash": base64.RawURLEncoding.EncodeToString(hash[:16])})
	calls := 0
	*f.exchange = func(w http.ResponseWriter, req *http.Request) {
		calls++
		req.ParseForm()
		challenge := sha256.Sum256([]byte(req.Form.Get("code_verifier")))
		if req.Form.Get("client_secret") != "" || req.Form.Get("client_id") != "gmwe" || req.Form.Get("resource") != f.a.Config.Resource || base64.RawURLEncoding.EncodeToString(challenge[:]) != q.Get("code_challenge") {
			t.Error("incorrect PKCE exchange")
		}
		json.NewEncoder(w).Encode(map[string]any{"access_token": access, "refresh_token": "fixture-refresh", "id_token": id, "token_type": "Bearer", "expires_in": 3600})
	}
	callbackURL := "https://gmwe.test/api/auth/callback?code=test&state=" + q.Get("state") + "&iss=" + url.QueryEscape(f.issuer)
	req := httptest.NewRequest("GET", callbackURL, nil)
	req.AddCookie(start.Result().Cookies()[0])
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != 303 || res.Header().Get("Location") != "/account" {
		t.Fatalf("callback failed: %d", res.Code)
	}
	cookie := res.Result().Cookies()[0]
	if cookie.Value == start.Result().Cookies()[0].Value {
		t.Fatal("session not rotated")
	}
	meReq := httptest.NewRequest("GET", "https://gmwe.test/api/auth/me", nil)
	meReq.AddCookie(cookie)
	me := httptest.NewRecorder()
	h.ServeHTTP(me, meReq)
	if me.Code != 200 || strings.Contains(me.Body.String(), access) || !strings.Contains(me.Body.String(), `"user_id":2`) {
		t.Fatal("invalid account response")
	}
	replay := httptest.NewRequest("GET", callbackURL, nil)
	replay.AddCookie(cookie)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, replay)
	if w.Code != 400 || calls != 1 {
		t.Fatal("callback replay accepted")
	}
}

func TestPrivatePageAndRetiredShortcut(t *testing.T) {
	f := newFixture(t)
	r := gin.New()
	f.a.Routes(r)
	r.GET("/seed", func(c *gin.Context) {
		f.a.Sessions.Put(c.Request.Context(), "access", f.token(t, "at+jwt", nil))
		c.Status(204)
	})
	h := f.a.Sessions.LoadAndSave(r)
	seed := httptest.NewRecorder()
	h.ServeHTTP(seed, httptest.NewRequest("GET", "https://gmwe.test/seed", nil))
	for _, signedIn := range []bool{false, true} {
		req := httptest.NewRequest("GET", "https://gmwe.test/api/auth/page-access", nil)
		if signedIn {
			req.AddCookie(seed.Result().Cookies()[0])
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if signedIn && w.Code != 204 {
			t.Fatal("session page denied")
		}
		if !signedIn && (w.Code != 302 || w.Header().Get("Location") != "/login") {
			t.Fatal("private page exposed")
		}
	}
	for _, path := range []string{"/api/auth/shortcut/start", "/api/auth/shortcut/finish"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", "https://gmwe.test"+path, nil))
		if w.Code != 404 {
			t.Fatal("retired pairing still registered")
		}
	}
}
