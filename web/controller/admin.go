package controller

import (
    "net/http"
    "strconv"

    "github.com/mhsanaei/3x-ui/v3/web/middleware"
    "github.com/mhsanaei/3x-ui/v3/web/service"
    "github.com/gin-gonic/gin"
)

type AdminController struct {
    BaseController
    userService service.UserService
}

func NewAdminController(g *gin.RouterGroup) *AdminController {
    a := &AdminController{}
    // All routes here already sit under /panel/api + checkAPIAuth.
    // We add RequireOwner on top.
    admins := g.Group("/admins")
    admins.Use(middleware.RequireOwner())
    admins.GET("", a.list)
    admins.POST("/create", a.create)
    admins.POST("/update/:id", a.update)
    admins.POST("/delete/:id", a.delete)
    return a
}

type subAdminForm struct {
    Username        string `json:"username" form:"username" binding:"required"`
    Password        string `json:"password" form:"password"`
    AllowedInbounds []int  `json:"allowedInbounds" form:"allowedInbounds"`
}

func (a *AdminController) list(c *gin.Context) {
    users, err := a.userService.ListSubAdmins()
    jsonObj(c, users, err)
}

func (a *AdminController) create(c *gin.Context) {
    var form subAdminForm
    if err := c.ShouldBind(&form); err != nil {
        pureJsonMsg(c, http.StatusOK, false, err.Error())
        return
    }
    if form.Password == "" {
        pureJsonMsg(c, http.StatusOK, false, "password is required")
        return
    }
    err := a.userService.CreateSubAdmin(form.Username, form.Password, form.AllowedInbounds)
    jsonMsg(c, "create admin", err)
}

func (a *AdminController) update(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    var form subAdminForm
    if err := c.ShouldBind(&form); err != nil {
        pureJsonMsg(c, http.StatusOK, false, err.Error())
        return
    }
    err := a.userService.UpdateSubAdmin(id, form.Username, form.Password, form.AllowedInbounds)
    jsonMsg(c, "update admin", err)
}

func (a *AdminController) delete(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    err := a.userService.DeleteSubAdmin(id)
    jsonMsg(c, "delete admin", err)
}