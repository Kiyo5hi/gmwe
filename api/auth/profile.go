package auth

import (
	"context"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type accountProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Loaded    bool   `json:"loaded"`
	State     string `json:"state"`
}

func (a *OIDCAuth) saveProfile(ctx context.Context, first, last *string) {
	for key, value := range map[string]*string{"first_name": first, "last_name": last} {
		a.Sessions.Remove(ctx, key)
		if value != nil {
			a.Sessions.Put(ctx, key, strings.TrimSpace(*value))
		}
	}
	a.Sessions.Put(ctx, "profile_loaded", true)
	a.Sessions.Put(ctx, "profile_state", "ready")
}

func (a *OIDCAuth) profile(ctx context.Context, writer Writer) accountProfile {
	read := func() accountProfile {
		return accountProfile{a.Sessions.GetString(ctx, "first_name"), a.Sessions.GetString(ctx, "last_name"),
			a.Sessions.GetBool(ctx, "profile_loaded"), a.Sessions.GetString(ctx, "profile_state")}
	}
	profile := read()
	if profile.State == "ready" || profile.State == "consent_required" {
		return profile
	}
	if profile.Loaded && (profile.FirstName != "" || profile.LastName != "") {
		a.Sessions.Put(ctx, "profile_state", "ready")
		return read()
	}
	// Backfill old sessions via the provider's standard endpoint, not its admin API.
	// Bound retries so a provider outage cannot delay every route transition.
	if time.Since(time.Unix(a.Sessions.GetInt64(ctx, "profile_attempt"), 0)) < time.Minute {
		return profile
	}
	a.Sessions.Put(ctx, "profile_attempt", time.Now().Unix())
	a.Sessions.Put(ctx, "profile_loaded", false)
	a.Sessions.Put(ctx, "profile_state", "unavailable")
	requestCtx, cancel := context.WithTimeout(a.context(ctx), 4*time.Second)
	defer cancel()
	info, err := a.provider.UserInfo(requestCtx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: a.Sessions.GetString(ctx, "access")}))
	if err != nil || info.Subject != writer.Subject {
		return read()
	}
	var claims struct {
		FirstName *string `json:"given_name"`
		LastName  *string `json:"family_name"`
	}
	if info.Claims(&claims) != nil {
		return read()
	}
	if claims.FirstName == nil && claims.LastName == nil {
		a.Sessions.Put(ctx, "profile_state", "consent_required")
		return read()
	}
	a.saveProfile(ctx, claims.FirstName, claims.LastName)
	return read()
}
