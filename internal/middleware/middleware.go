package middleware

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func StaticFileMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apply Content-Type based on file extension for static files under /apps/
		// This prevents mislabeling non-APK files (e.g., sbom.json) as APK.
		if strings.HasPrefix(c.Request.URL.Path, "/apps/") {
			ext := strings.ToLower(filepath.Ext(c.Request.URL.Path))
			switch ext {
			case ".apk":
				c.Header("Content-Type", "application/vnd.android.package-archive")
			case ".aab":
				// AAB is a ZIP-based bundle
				c.Header("Content-Type", "application/zip")
			case ".json":
				c.Header("Content-Type", "application/json")
			default:
				// Let the default content type be determined by the server
			}
		}
		// Continue to the next middleware or handler
		c.Next()
	}
}

/*
import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sorokinmax/websspi"
)

const UserInfoKey = "websspi-key-UserInfo"

func AddUserToCtx() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ctxVars, ok := c.Request.Context().Value(UserInfoKey).(*websspi.UserInfo); ok {
			c.Set("user", ctxVars.Username)
		} else {
			//c.Set("user", "guest")
			c.Abort()
			//c.Next()
			return
		}
	}
}

func MidAuth(a *websspi.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {

		user, data, err := a.Authenticate(c.Request, c.Writer)
		if err != nil {
			a.Return401(c.Writer, data)
			return
		}

		// Add the UserInfo value to the reqest's context
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), UserInfoKey, user))
		// and to the request header with key Config.AuthUserKey
		if a.Config.AuthUserKey != "" {
			c.Request.Header.Set(a.Config.AuthUserKey, user.Username)
		}

		// The WWW-Authenticate header might need to be sent back even
		// on successful authentication (eg. in order to let the client complete
		// mutual authentication).
		if data != "" {
			a.AppendAuthenticateHeader(c.Writer, data)
		}

		c.Next()
	}
}
*/
