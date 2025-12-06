package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wuzhengjerry/udns-api/frame"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/role"
	"github.com/wuzhengjerry/udns-api/pkg/route"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateRoute 新增线路
// @Summary 新增线路
// @version 1.0.0
// @tags 线路管理
// @description  传递参数新增线路
// @Produce json
// @Param StaffName header string true "用户名"
// @Param object body route.Route true "新增线路参数"
// @Success 200 {object} frame.AuthResponse
// @Router /route [post]
func (h *handler) CreateRoute(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	r := route.Route{}
	err := c.BindJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	if err = r.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	err = h.service.InsertRoute(&r)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse("create route success"))
}

// DeleteRoute 删除线路
// @Summary 删除线路
// @version 1.0.0
// @tags 线路管理
// @description  根据ID删除具体线路
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param id path string true "线路 ID"
// @Success 200 {object} frame.AuthResponse
// @Router /route/{id} [delete]
func (h *handler) DeleteRoute(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	id := c.Param("id")

	err := h.service.DeleteRoute(id)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}

	c.JSON(http.StatusOK, frame.NewResponse("delete route success"))
}

// UpdateRoute 编辑线路
// @Summary 编辑线路
// @version 1.0.0
// @tags 线路管理
// @description  根据ID编辑具体线路
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param object body route.Route true "编辑线路参数"
// @Param id path string true "线路 ID"
// @Success 200 {object} frame.AuthResponse
// @Router /route/{id} [put]
func (h *handler) UpdateRoute(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	r := route.Route{}
	err := c.BindJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	r.ID, _ = primitive.ObjectIDFromHex(c.Param("id"))
	err = h.service.UpdateRoute(&r)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}

	c.JSON(http.StatusOK, frame.NewResponse("update route success"))
}

// QueryRoutes 线路列表
// @Summary 线路列表
// @version 1.0.0
// @tags 线路管理
// @description  根据名称匹配查询线路列表
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param ps query string true "page size"
// @Param pn query string true "page number"
// @Param name query string false "基础域关键字，支持模糊匹配，可为空"
// @Success 200 {object} frame.AuthResponse
// @Router /route [get]
func (h *handler) QueryRoutes(c *gin.Context) {

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

	count, a, err := h.service.QueryRoutes(name, int64(ps), int64(pn))
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, frame.NewPagedResponse(a, ps, pn, int(count)))
}
