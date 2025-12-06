package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wuzhengjerry/udns-api/frame"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/role"
	"github.com/wuzhengjerry/udns-api/pkg/webuser"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateUser 创建用户
func (h *handler) CreateUser(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	z := webuser.WebUser{}
	err := c.BindJSON(&z)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	if err = z.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	err = h.service.InsertUser(&z)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse("create user success"))
}

// DeleteUser 删除用户
func (h *handler) DeleteUser(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	name := c.Param("name")
	err := h.service.DeleteUser(name)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}

	c.JSON(http.StatusOK, frame.NewResponse("delete user success"))
}

// UpdateUser 更新用户
func (h *handler) UpdateUser(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	z := webuser.WebUser{}
	err := c.BindJSON(&z)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	z.ID, _ = primitive.ObjectIDFromHex(c.Param("id"))
	err = h.service.UpdateUser(&z)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}

	c.JSON(http.StatusOK, frame.NewResponse("update user success"))
}

// QueryUsers 查询用户
func (h *handler) QueryUsers(c *gin.Context) {

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

	count, a, err := h.service.QueryUsers(name, int64(ps), int64(pn))
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, frame.NewPagedResponse(a, ps, pn, int(count)))
}

// DescribeUser 用户详情
func (h *handler) DescribeUser(c *gin.Context) {
	name := c.Param("name")
	res, err := h.service.DescribeUser(name)
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
