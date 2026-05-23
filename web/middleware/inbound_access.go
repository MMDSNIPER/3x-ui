package middleware

import (
	"net/http"

	"github.com/mhsanaei/3x-ui/v3/web/session"

	"github.com/gin-gonic/gin"
)

func RequireOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := session.GetLoginUser(c)
		if user == nil || !user.IsOwner() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false, "msg": "owner access required",
			})
			return
		}
		c.Next()
	}
}

func InjectInboundAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := session.GetLoginUser(c)
		if user != nil {
			c.Set("allowed_inbounds", user.GetAllowedInboundIDs())
			c.Set("user_has_full_access", user.HasFullAccess())
		}
		c.Next()
	}
}

func AllowedInboundsFromCtx(c *gin.Context) ([]int, bool) {
	full, _ := c.Get("user_has_full_access")
	if full == true {
		return nil, true
	}
	ids, _ := c.Get("allowed_inbounds")
	if v, ok := ids.([]int); ok {
		return v, false
	}
	return nil, true
}