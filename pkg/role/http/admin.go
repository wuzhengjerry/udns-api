package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wuzhengjerry/udns-api/frame"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/role"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateRole 编辑角色
// @Summary 编辑角色
// @version 1.0.0
// @tags 权限管理
// @description  根据ID编辑具体角色对应用户
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param object body role.Role true "编辑线路参数"
// @Param id path string true "角色 ID"
// @Success 200 {object} frame.AuthResponse
// @Router /role/{id} [put]
func (h *handler) UpdateRole(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	r := role.Role{}
	err := c.BindJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	r.ID, _ = primitive.ObjectIDFromHex(c.Param("id"))
	err = h.service.UpdateRole(&r)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}

	c.JSON(http.StatusOK, frame.NewResponse("update Role success"))
}

// QueryRoles 角色列表
// @Summary 角色列表
// @version 1.0.0
// @tags 权限管理
// @description  查询角色列表
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param ps query string true "page size"
// @Param pn query string true "page number"
// @Param name query string false "基础域关键字，支持模糊匹配，可为空"
// @Success 200 {object} frame.AuthResponse
// @Router /role [get]
func (h *handler) QueryRoles(c *gin.Context) {
	ps, err := strconv.Atoi(c.Query("ps"))
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	pn, err := strconv.Atoi(c.Query("pn"))
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	name := c.Query("name")

	count, a, err := h.service.QueryRoles(name, int64(ps), int64(pn))
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, frame.NewPagedResponse(a, ps, pn, int(count)))
}
