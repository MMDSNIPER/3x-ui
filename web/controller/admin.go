// FILE: web/controller/admin.go  (new file)

package controller

import (
	"net/http"
	"strconv"

	"github.com/mhsanaei/3x-ui/v3/web/service"
	"github.com/mhsanaei/3x-ui/v3/web/session"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	BaseController
	userService service.UserService
}

func NewAdminController(g *gin.RouterGroup) *AdminController {
	a := &AdminController{}
	a.initRouter(g)
	return a
}

type subAdminForm struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	AllowedInbounds []int  `json:"allowedInbounds"`
}

func (a *AdminController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/admins")
	g.GET("/me", a.me)
	g.Use(a.requireMainAdmin)
	g.GET("", a.getAdmins)
	g.POST("", a.createAdmin)
	g.PUT("/:id", a.updateAdmin)
	g.DELETE("/:id", a.deleteAdmin)
}

func (a *AdminController) requireMainAdmin(c *gin.Context) {
	user := session.GetLoginUser(c)
	if user == nil || (user.Role != "admin" && user.Role != "") {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "msg": "Forbidden"})
		return
	}
	c.Next()
}

func (a *AdminController) me(c *gin.Context) {
	user := session.GetLoginUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
		return
	}
	allowedIDs := a.userService.GetAllowedInboundIDs(user)
	jsonObj(c, gin.H{
		"id":              user.Id,
		"username":        user.Username,
		"role":            user.Role,
		"allowedInbounds": allowedIDs,
	}, nil)
}

func (a *AdminController) getAdmins(c *gin.Context) {
	admins, err := a.userService.GetAllSubAdmins()
	if err != nil {
		jsonMsg(c, "failed to get admins", err)
		return
	}
	jsonObj(c, admins, nil)
}

func (a *AdminController) createAdmin(c *gin.Context) {
	var form subAdminForm
	if err := c.ShouldBindJSON(&form); err != nil {
		jsonMsg(c, "invalid form", err)
		return
	}
	if form.Username == "" || form.Password == "" {
		jsonMsg(c, "username and password required", nil)
		return
	}
	admin, err := a.userService.CreateSubAdmin(form.Username, form.Password, form.AllowedInbounds)
	if err != nil {
		jsonMsg(c, "failed to create admin", err)
		return
	}
	jsonObj(c, admin, nil)
}

func (a *AdminController) updateAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "invalid id", err)
		return
	}
	var form subAdminForm
	if err := c.ShouldBindJSON(&form); err != nil {
		jsonMsg(c, "invalid form", err)
		return
	}
	if err := a.userService.UpdateSubAdmin(id, form.Username, form.Password, form.AllowedInbounds); err != nil {
		jsonMsg(c, "failed to update admin", err)
		return
	}
	jsonMsg(c, "updated", nil)
}

func (a *AdminController) deleteAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "invalid id", err)
		return
	}
	if err := a.userService.DeleteSubAdmin(id); err != nil {
		jsonMsg(c, "failed to delete admin", err)
		return
	}
	jsonMsg(c, "deleted", nil)
}