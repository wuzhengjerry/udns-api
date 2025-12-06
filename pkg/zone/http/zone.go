package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wuzhengjerry/udns-api/frame"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/role"
	"github.com/wuzhengjerry/udns-api/pkg/zone"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateZone 新增二级zone
// @Summary 新增二级zone
// @version 1.0.0
// @tags 二级zone管理
// @description  传递参数新增二级zone
// @Produce json
// @Param StaffName header string true "用户名"
// @Param object body zone.Zone true "新增二级zone参数"
// @Success 200 {object} frame.AuthResponse
// @Router /zone [post]
func (h *handler) CreateZone(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	z := zone.Zone{}
	err := c.BindJSON(&z)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	if err = z.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	err = h.service.InsertZone(&z)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse("create zone success"))
}

// DeleteZone  删除二级zone
// @Summary 删除二级zone
// @version 1.0.0
// @tags 二级zone管理
// @description  根据ID删除具体二级zone
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param id path string true "二级zone ID"
// @Success 200 {object} frame.AuthResponse
// @Router /zone/{id} [delete]
func (h *handler) DeleteZone(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	id := c.Param("id")

	err := h.service.DeleteZone(id)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}

	c.JSON(http.StatusOK, frame.NewResponse("delete zone success"))
}

// UpdateZone 编辑二级zone
// @Summary 编辑二级zone
// @version 1.0.0
// @tags 二级zone管理
// @description  根据ID编辑具体二级zone
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param object body zone.Zone true "编辑二级zone参数"
// @Param id path string true "二级zone ID"
// @Success 200 {object} frame.AuthResponse
// @Router /zone/{id} [put]
func (h *handler) UpdateZone(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	z := zone.Zone{}
	err := c.BindJSON(&z)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	z.ID, _ = primitive.ObjectIDFromHex(c.Param("id"))
	err = h.service.UpdateZone(&z)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}

	c.JSON(http.StatusOK, frame.NewResponse("update zone success"))
}

// QueryZones 二级zone列表
// @Summary 二级zone列表
// @version 1.0.0
// @tags 二级zone管理
// @description  根据名称匹配查询二级zone列表
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param ps query string true "page size"
// @Param pn query string true "page number"
// @Param name query string false "基础域关键字，支持模糊匹配，可为空"
// @Success 200 {object} frame.AuthResponse
// @Router /zone [get]
func (h *handler) QueryZones(c *gin.Context) {

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

	count, a, err := h.service.QueryZones(name, int64(ps), int64(pn))
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, frame.NewPagedResponse(a, ps, pn, int(count)))
}

// DescribeZone 查询zone
func (h *handler) DescribeZone(c *gin.Context) {
	name := c.Param("name")
	res, err := h.service.DescribeZone(name)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusOK, frame.NewInternalError(err))
			return
		}
	} else {
		c.JSON(http.StatusOK, frame.NewResponse(res))
	}

}
