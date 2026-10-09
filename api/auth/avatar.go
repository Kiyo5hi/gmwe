package auth

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Pocket ID exposes public avatars by subject. Resolve only server-configured
// members; never accept a client-supplied subject or destination URL.
func (a *OIDCAuth) avatar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		c.AbortWithStatus(404)
		return
	}
	for _, member := range a.Config.Writers {
		if member.UserID != id {
			continue
		}
		destination := strings.TrimRight(a.Config.Issuer, "/") + "/api/users/" + url.PathEscape(member.Subject) + "/profile-picture.png"
		// A bounded cache key prevents stale CDN initials persisting after IdP edits.
		destination += "?v=" + strconv.FormatInt(time.Now().Unix()/300, 10)
		c.Header("Cache-Control", "private, no-store")
		c.Header("Referrer-Policy", "no-referrer")
		c.Redirect(http.StatusFound, destination)
		return
	}
	c.AbortWithStatus(404)
}
