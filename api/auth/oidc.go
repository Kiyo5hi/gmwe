package auth

import (
	"context"
	"crypto/subtle"
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	jose "github.com/go-jose/go-jose/v4"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

const WriteScope = "write:hitokoto"

//go:embed login.html
var loginHTML string

type Writer struct {
	Subject string `json:"subject"`
	UserID  int    `json:"user_id"`
	Name    string `json:"name"`
}

type OIDCConfig struct {
	Issuer   string   `json:"issuer"`
	ClientID string   `json:"client_id"`
	Resource string   `json:"resource"`
	Origin   string   `json:"origin"`
	Writers  []Writer `json:"writers"`
}

func LoadOIDCConfig(path string) (OIDCConfig, error) {
	var cfg OIDCConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, errors.New("OIDC configuration unavailable")
	}
	if err = json.Unmarshal(data, &cfg); err != nil {
		return cfg, errors.New("invalid OIDC configuration")
	}
	for _, value := range []string{cfg.Issuer, cfg.Origin, cfg.Resource} {
		u, err := url.Parse(value)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return cfg, errors.New("OIDC requires explicit HTTPS URLs")
		}
	}
	u, _ := url.Parse(cfg.Origin)
	if u.Path != "" || cfg.ClientID == "" || cfg.Resource == cfg.ClientID || len(cfg.Writers) == 0 {
		return cfg, errors.New("invalid OIDC identity boundary")
	}
	subjects, users := map[string]bool{}, map[int]bool{}
	for _, w := range cfg.Writers {
		if w.Subject == "" || w.UserID <= 0 || subjects[w.Subject] || users[w.UserID] {
			return cfg, errors.New("ambiguous OIDC writer mapping")
		}
		subjects[w.Subject], users[w.UserID] = true, true
	}
	return cfg, nil
}

type OIDCAuth struct {
	Config       OIDCConfig
	Sessions     *scs.SessionManager
	client       *oauth2.Config
	idVerifier   *oidc.IDTokenVerifier
	provider     *oidc.Provider
	apiVerifier  *oidc.IDTokenVerifier
	httpClient   *http.Client
	loginLimit   *rate.Limiter
	sessionLocks [64]sync.Mutex
}

func NewOIDC(ctx context.Context, cfg OIDCConfig) (*OIDCAuth, error) {
	httpClient := &http.Client{Timeout: 15 * time.Second}
	ctx = oidc.ClientContext(ctx, httpClient)
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, errors.New("OIDC discovery failed")
	}
	sessions := scs.New()
	sessions.Lifetime = sessionLifetime
	sessions.IdleTimeout = 7 * 24 * time.Hour
	sessions.Cookie.Name = "__Host-gmwe"
	sessions.Cookie.Secure = true
	sessions.Cookie.HttpOnly = true
	sessions.Cookie.SameSite = http.SameSiteLaxMode
	sessions.Cookie.Persist = true
	sessions.HashTokenInStore = true
	sessions.ErrorFunc = func(w http.ResponseWriter, r *http.Request, err error) { http.Error(w, "Session unavailable", 503) }
	endpoint := provider.Endpoint()
	endpoint.AuthStyle = oauth2.AuthStyleInParams
	return &OIDCAuth{Config: cfg, Sessions: sessions, provider: provider, httpClient: httpClient, loginLimit: rate.NewLimiter(rate.Every(10*time.Second), 10),
		client: &oauth2.Config{ClientID: cfg.ClientID, Endpoint: endpoint,
			RedirectURL: cfg.Origin + "/api/auth/callback", Scopes: []string{oidc.ScopeOpenID, "profile", WriteScope, oidc.ScopeOfflineAccess}},
		idVerifier:  provider.Verifier(&oidc.Config{ClientID: cfg.ClientID, SupportedSigningAlgs: []string{"RS256"}}),
		apiVerifier: provider.Verifier(&oidc.Config{ClientID: cfg.Resource, SupportedSigningAlgs: []string{"RS256"}})}, nil
}

func (a *OIDCAuth) context(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, a.httpClient)
}

func (a *OIDCAuth) VerifyAccess(ctx context.Context, raw string) (Writer, error) {
	denied := errors.New("invalid API credential")
	if len(raw) == 0 || len(raw) > 16384 {
		return Writer{}, denied
	}
	parsed, err := jose.ParseSigned(raw, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil || len(parsed.Signatures) != 1 || parsed.Signatures[0].Protected.ExtraHeaders[jose.HeaderType] != "at+jwt" {
		return Writer{}, denied
	}
	token, err := a.apiVerifier.Verify(a.context(ctx), raw)
	if err != nil {
		return Writer{}, denied
	}
	var claims struct {
		Scope     string `json:"scope"`
		ClientID  string `json:"client_id"`
		IssuedAt  int64  `json:"iat"`
		NotBefore int64  `json:"nbf"`
	}
	if token.Claims(&claims) != nil || claims.ClientID != a.Config.ClientID ||
		!slices.Contains(strings.Fields(claims.Scope), WriteScope) || claims.IssuedAt <= 0 ||
		claims.IssuedAt > time.Now().Add(time.Minute).Unix() || claims.NotBefore > time.Now().Unix() {
		return Writer{}, denied
	}
	for _, writer := range a.Config.Writers {
		if writer.Subject == token.Subject {
			return writer, nil
		}
	}
	return Writer{}, denied
}

func (a *OIDCAuth) csrf(c *gin.Context) bool {
	expected := a.Sessions.GetString(c.Request.Context(), "csrf")
	return c.GetHeader("Origin") == a.Config.Origin && expected != "" &&
		subtle.ConstantTimeCompare([]byte(expected), []byte(c.GetHeader("X-CSRF-Token"))) == 1
}

func (a *OIDCAuth) RequireWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if a == nil {
			c.AbortWithStatus(401)
			return
		}
		c.Header("Cache-Control", "private, no-store")
		raw := a.Sessions.GetString(c.Request.Context(), "access")
		if c.GetHeader("Authorization") != "" || raw == "" {
			c.AbortWithStatus(401)
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && !a.csrf(c) {
			c.AbortWithStatus(403)
			return
		}
		writer, err := a.sessionWriter(c.Request.Context())
		if err != nil {
			if errors.Is(err, errSessionUnavailable) {
				c.AbortWithStatus(503)
				return
			}
			c.AbortWithStatus(401)
			return
		}
		c.Set("writer", writer)
		c.Next()
	}
}

func (a *OIDCAuth) Routes(r *gin.Engine) {
	r.GET("/login", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Data(200, "text/html; charset=utf-8", []byte(loginHTML))
	})
	group := r.Group("/api/auth")
	group.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	})
	group.GET("/login", a.login)
	group.GET("/callback", a.callback)
	group.GET("/page-access", func(c *gin.Context) {
		if c.GetHeader("Authorization") != "" {
			c.Redirect(302, "/login")
			return
		}
		_, err := a.sessionWriter(c.Request.Context())
		if errors.Is(err, errSessionUnavailable) {
			c.AbortWithStatus(503)
			return
		}
		if err != nil {
			c.Redirect(302, "/login")
			return
		}
		c.Status(204)
	})
	group.GET("/me", a.RequireWriter(), func(c *gin.Context) {
		ctx := c.Request.Context()
		c.JSON(200, gin.H{"user": c.MustGet("writer"), "csrf": a.Sessions.GetString(ctx, "csrf"),
			"profile": a.profile(ctx, c.MustGet("writer").(Writer))})
	})
	group.POST("/logout", a.RequireWriter(), func(c *gin.Context) {
		if err := a.Sessions.Destroy(c.Request.Context()); err != nil {
			c.AbortWithStatus(500)
			return
		}
		c.Status(204)
	})
}

func (a *OIDCAuth) login(c *gin.Context) {
	if !a.loginLimit.Allow() {
		c.Header("Retry-After", "10")
		c.AbortWithStatus(429)
		return
	}
	c.Redirect(302, a.authorizationURL(c.Request.Context()))
}

func (a *OIDCAuth) authorizationURL(ctx context.Context) string {
	// SCS keeps PKCE/state server-side; only an opaque host-only cookie reaches the browser.
	state, nonce, verifier := oauth2.GenerateVerifier(), oauth2.GenerateVerifier(), oauth2.GenerateVerifier()
	a.Sessions.Put(ctx, "state", state)
	a.Sessions.Put(ctx, "nonce", nonce)
	a.Sessions.Put(ctx, "verifier", verifier)
	a.Sessions.Put(ctx, "login_time", time.Now().Unix())
	a.Sessions.SetDeadline(ctx, time.Now().Add(10*time.Minute))
	options := []oauth2.AuthCodeOption{oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("resource", a.Config.Resource), oauth2.SetAuthURLParam("prompt", "consent")}
	return a.client.AuthCodeURL(state, options...)
}

func (a *OIDCAuth) callback(c *gin.Context) {
	ctx := c.Request.Context()
	state := a.Sessions.PopString(ctx, "state")
	nonce := a.Sessions.PopString(ctx, "nonce")
	verifier := a.Sessions.PopString(ctx, "verifier")
	started := time.Unix(a.Sessions.GetInt64(ctx, "login_time"), 0)
	a.Sessions.Remove(ctx, "login_time")
	if state == "" || nonce == "" || verifier == "" || time.Since(started) > 10*time.Minute ||
		subtle.ConstantTimeCompare([]byte(state), []byte(c.Query("state"))) != 1 ||
		c.Query("iss") != a.Config.Issuer || c.Query("code") == "" || c.Query("error") != "" {
		c.AbortWithStatus(400)
		return
	}
	token, err := a.client.Exchange(a.context(ctx), c.Query("code"), oauth2.VerifierOption(verifier),
		oauth2.SetAuthURLParam("resource", a.Config.Resource))
	if err != nil {
		c.AbortWithStatus(401)
		return
	}
	rawID, ok := token.Extra("id_token").(string)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	id, err := a.idVerifier.Verify(a.context(ctx), rawID)
	if err != nil || id.Nonce != nonce || id.VerifyAccessToken(token.AccessToken) != nil {
		c.AbortWithStatus(401)
		return
	}
	writer, err := a.VerifyAccess(ctx, token.AccessToken)
	if err != nil || writer.Subject != id.Subject || token.RefreshToken == "" {
		c.AbortWithStatus(403)
		return
	}
	if err := a.Sessions.RenewToken(ctx); err != nil {
		c.AbortWithStatus(500)
		return
	}
	var profile struct {
		FirstName *string `json:"given_name"`
		LastName  *string `json:"family_name"`
	}
	if id.Claims(&profile) != nil {
		c.AbortWithStatus(401)
		return
	}
	a.Sessions.Remove(ctx, "first_name")
	a.Sessions.Remove(ctx, "last_name")
	a.Sessions.Remove(ctx, "profile_state")
	a.Sessions.Remove(ctx, "profile_attempt")
	a.Sessions.Put(ctx, "profile_loaded", false)
	if profile.FirstName != nil || profile.LastName != nil {
		a.saveProfile(ctx, profile.FirstName, profile.LastName)
	}
	a.Sessions.Put(ctx, "access", token.AccessToken)
	a.Sessions.Put(ctx, "refresh", token.RefreshToken)
	a.Sessions.Put(ctx, "subject", writer.Subject)
	a.Sessions.SetDeadline(ctx, time.Now().Add(sessionLifetime))
	a.Sessions.Put(ctx, "csrf", oauth2.GenerateVerifier())
	c.Redirect(303, "/account")
}
