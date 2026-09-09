package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
)

var errSessionUnavailable = errors.New("session renewal unavailable")

// PersistSessions keeps authentication separate from the application's content DB.
func (a *OIDCAuth) PersistSessions(path string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	f.Close()
	if err := os.Chmod(path, 0600); err != nil {
		return nil, err
	}
	uriPath := filepath.ToSlash(path)
	if filepath.VolumeName(path) != "" {
		uriPath = "/" + uriPath
	}
	u := url.URL{Scheme: "file", Path: uriPath}
	database, err := sql.Open("sqlite3", u.String()+"?_busy_timeout=5000&_journal_mode=DELETE&_synchronous=FULL")
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(1)
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS sessions (token TEXT PRIMARY KEY, data BLOB NOT NULL, expiry REAL NOT NULL); CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry)`)
	if err != nil {
		database.Close()
		return nil, err
	}
	store := sqlite3store.New(database)
	a.Sessions.Store = store
	return func() { store.StopCleanup(); database.Close() }, nil
}

// Lock before LoadAndSave: concurrent requests must see the committed rotated token.
// The application intentionally runs a single instance with local SQLite.
func (a *OIDCAuth) Middleware(next http.Handler) http.Handler {
	handler := a.Sessions.LoadAndSave(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(a.Sessions.Cookie.Name); err == nil {
			hash := sha256.Sum256([]byte(cookie.Value))
			lock := &a.sessionLocks[int(hash[0])%len(a.sessionLocks)]
			lock.Lock()
			defer lock.Unlock()
		}
		handler.ServeHTTP(w, r)
	})
}

func (a *OIDCAuth) sessionWriter(ctx context.Context) (Writer, error) {
	raw := a.Sessions.GetString(ctx, "access")
	writer, err := a.VerifyAccess(ctx, raw)
	if err == nil {
		return writer, nil
	}
	refresh := a.Sessions.GetString(ctx, "refresh")
	subject := a.Sessions.GetString(ctx, "subject")
	if refresh == "" || subject == "" {
		return Writer{}, errors.New("login required")
	}
	// Revalidate provider signatures, audience, client, scope and subject after renewal.
	token, err := a.client.TokenSource(a.context(ctx), &oauth2.Token{RefreshToken: refresh}).Token()
	if err != nil {
		var rejection *oauth2.RetrieveError
		if errors.As(err, &rejection) && rejection.ErrorCode == "invalid_grant" {
			if a.Sessions.Destroy(ctx) != nil {
				return Writer{}, errSessionUnavailable
			}
			return Writer{}, errors.New("login required")
		}
		return Writer{}, errSessionUnavailable
	}
	writer, err = a.VerifyAccess(ctx, token.AccessToken)
	if err != nil || writer.Subject != subject || token.RefreshToken == "" {
		if a.Sessions.Destroy(ctx) != nil {
			return Writer{}, errSessionUnavailable
		}
		return Writer{}, errors.New("login required")
	}
	a.Sessions.Put(ctx, "access", token.AccessToken)
	a.Sessions.Put(ctx, "refresh", token.RefreshToken)
	return writer, nil
}

const sessionLifetime = 30 * 24 * time.Hour
