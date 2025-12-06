package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wuzhengjerry/udns-api/frame"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/domain"
	"github.com/wuzhengjerry/udns-api/pkg/role"
)

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

// CreateDomain 创建域名
func (h *handler) CreateDomain(c *gin.Context) {
	user := c.GetHeader("StaffName")
	d := domain.Domain{}
	err := c.BindJSON(&d)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	if err = d.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	err = h.service.InsertDomain(user, &d)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse("create Domain success"))
}

// DeleteDomain 删除域名
func (h *handler) DeleteDomain(c *gin.Context) {
	name := c.Param("name")
	user := c.GetHeader("StaffName")
	err := h.service.DeleteDomain(user, name)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}
	c.JSON(http.StatusOK, frame.NewResponse("delete remain success"))
}

// UpdateDomain 更新域名
func (h *handler) UpdateDomain(c *gin.Context) {
	// 校验用户权限
	user := c.GetHeader("StaffName")
	r := domain.Domain{}
	err := c.BindJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}

	flag := false
	if pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		flag = true
	}
	r.Name = c.Param("name")
	if h.service.HasDomainPermission(r.Name, user) {
		flag = true
	}

	if !flag {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	err = h.service.UpdateDomain(user, &r) //更新域名责任人
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}
	c.JSON(http.StatusOK, frame.NewResponse("update Domain success"))
}

// DescribeDomain 查询域名信息
func (h *handler) DescribeDomain(c *gin.Context) {
	name := c.Param("name")
	res, err := h.service.DescribeDomain(name)
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

// QueryDomains 查询域名列表
func (h *handler) QueryDomains(c *gin.Context) {
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
	manage := c.Query("manage")
	name := c.Query("name")
	owner := c.Query("owner")
	operator := c.Query("operator")
	description := c.Query("description")
	businessId := c.Query("business_id")

	count, a, err := h.service.QueryDomains(manage, name, owner, operator, description, businessId, int64(ps), int64(pn))
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, frame.NewPagedResponse(a, ps, pn, int(count)))
}

// QuerySyncApiLogs 查询日志列表
func (h *handler) QuerySyncApiLogs(c *gin.Context) {
	ps, err := strconv.Atoi(c.Query("ps"))
	//user := c.GetHeader("StaffName")
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	pn, err := strconv.Atoi(c.Query("pn"))
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	name := c.Query("domain_name")
	method := c.Query("method")
	success := c.Query("success")
	user := c.Query("operator")
	apiType := c.Query("api_type")

	count, a, err := h.service.QuerySyncApiLogs(name, user, method, apiType, success, int64(ps), int64(pn))
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewPagedResponse(a, ps, pn, int(count)))
}

// CreateDomainLog 创建日志
func (h *handler) CreateDomainLog(c *gin.Context) {
	user := c.GetHeader("StaffName")
	d := domain.DomainLog{}
	err := c.BindJSON(&d)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}

	err = h.service.InsertDomainLog(user, &d)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse("create Domain success"))

}

// DeleteDomainLog 删除域名日志
func (h *handler) DeleteDomainLog(c *gin.Context) {
	id := c.Param("id")
	err := h.service.DeleteDomainLog(id)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}

	c.JSON(http.StatusOK, frame.NewResponse("delete remain success"))
}

// DescribeDomainLog 查询日志信息
func (h *handler) DescribeDomainLog(c *gin.Context) {
	id := c.Param("id")
	res, err := h.service.DescribeDomainLog(id)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, frame.NewNotFound(err))
			return
		} else {
			c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
			return
		}
	}

	c.JSON(http.StatusOK, frame.NewResponse(res))
}

// QueryDomainLogs 查询域名日志列表
func (h *handler) QueryDomainLogs(c *gin.Context) {
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

	count, a, err := h.service.QueryDomainLogs(name, int64(ps), int64(pn))
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, frame.NewPagedResponse(a, ps, pn, int(count)))
}

// GetAllStaffFullName 获取用户名称
func (h *handler) GetAllStaffFullName(c *gin.Context) {
	//user := c.GetHeader("StaffName")
	a, err := h.service.GetAllStaffFullName()
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))

}

// SyncITApi 数据同步到IT
func (h *handler) SyncITApi(c *gin.Context) { //获取域名配置
	params := domain.ITParam{}
	user := c.GetHeader("StaffName")
	err := c.BindJSON(&params)
	a, err := h.service.SyncITApi(user, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))

}

// GetDomainConfig 获取域名配置
func (h *handler) GetDomainConfig(c *gin.Context) { //获取域名配置
	//params := new(interface{})
	params := domain.QueryDomainParam{}
	err := c.BindJSON(&params)
	a, err := h.service.GetDomainConfig(&params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))

}

// DeleteDomainConfig 删除域名配置
func (h *handler) DeleteDomainConfig(c *gin.Context) { //获取域名配置
	user := c.GetHeader("StaffName")
	params := domain.DelDomainParam{}
	err := c.BindJSON(&params)
	a, err := h.service.DeleteDomainConfig(user, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))

}

// ModifyDomainConfig 修改域名配置
func (h *handler) ModifyDomainConfig(c *gin.Context) { //获取域名配置
	user := c.GetHeader("StaffName")
	params := domain.ModDomainParam{}
	err := c.BindJSON(&params)
	a, err := h.service.ModifyDomainConfig(user, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// AddDomainConfig 创建域名配置
func (h *handler) AddDomainConfig(c *gin.Context) { //获取域名配置
	params := domain.ModDomainParam{}
	err := c.BindJSON(&params)
	a, err := h.service.AddDomainConfig(&params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))

}

// GetStaffInfo 获取员工信息
func (h *handler) GetStaffInfo(c *gin.Context) {
	user := c.GetHeader("StaffName")
	userName := c.Query("user") // 有用户参数返回用户信息，否则返回查询人信息
	if userName == "" {
		userName = user
	}
	a, err := h.service.GetStaffInfo(userName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))

}

// GetITData 获取IT侧域名信息
func (h *handler) GetITData(c *gin.Context) {
	//user := c.GetHeader("StaffName")
	domainName := c.Query("domain") // 有用户参数返回用户信息，否则返回查询人信息

	itConfig, itErr := h.service.RequestRITApi(domainName)
	if itErr != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(itErr))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(itConfig))

}

// GetBusinessTree 获取用户组织信息
func (h *handler) GetBusinessTree(c *gin.Context) {
	//user := c.GetHeader("StaffName")
	a, err := h.service.GetBusinessTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))

}

// IsDomainsExisted 判断域名是否存在
func (h *handler) IsDomainsExisted(c *gin.Context) {
	name := c.Query("name")
	res, err := h.service.IsDomainsExisted(name)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(res))
}

// QueryITDomainsExisted 判断域名在IT侧是否存在
func (h *handler) QueryITDomainsExisted(c *gin.Context) {
	user := c.GetHeader("StaffName")
	name := c.Query("domain")
	res, err := h.service.QueryITDomainsExisted(user, name)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(res))
}
