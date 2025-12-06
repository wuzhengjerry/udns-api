package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wuzhengjerry/udns-api/frame"
	"github.com/wuzhengjerry/udns-api/pkg/openapi"
)

// QueryDomain 查询域名
func (h *handler) QueryDomain(c *gin.Context) {
	name := c.GetHeader("x-client-id")
	if name == "" {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("get user failed")))
		return
	}
	flag, domains, permError := h.service.IsUserValidate(name)
	if !flag {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("user permission deny")))
		return
	} else {
		if permError != nil {
			c.JSON(http.StatusBadRequest, frame.NewBadRequest(permError))
			return
		} else {
			d := openapi.QueryParam{Limit: 10, Offset: 0}
			err := c.BindJSON(&d)
			if err != nil {
				c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
				return
			}

			d.User = name
			d.Domains = domains
			limit := d.Limit
			offset := d.Offset
			count, a, rErr := h.service.QueryDomain(&d)
			if rErr != nil {
				c.JSON(http.StatusOK, frame.NewInternalError(rErr))
				return
			}
			c.JSON(http.StatusOK, frame.NewOpenApiResponse(a, int(limit), int(offset), int(count)))
		}
	}

}

// AddDomain 新增域名
func (h *handler) AddDomain(c *gin.Context) {
	name := c.GetHeader("x-client-id")
	if name == "" {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("get user failed")))
		return
	}
	flag, domains, permError := h.service.IsUserValidate(name)
	if permError != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(permError))
		return
	} else {
		if !flag {
			c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("user permission deny")))
			return
		} else {
			var d []*openapi.DomainInfo
			err := c.BindJSON(&d)
			if err != nil {
				c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
				return
			}

			res, err := h.service.AddDomain(d, name, domains)
			if err != nil {
				c.JSON(http.StatusOK, frame.NewOpenApiInternalError(res, err))
				return
			}
			c.JSON(http.StatusOK, frame.NewResponse(res))
		}
	}
}

// ModDomain 修改域名
func (h *handler) ModDomain(c *gin.Context) {
	name := c.GetHeader("x-client-id")
	if name == "" {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("get user failed")))
		return
	}
	flag, domains, permError := h.service.IsUserValidate(name)
	if permError != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(permError))
		return
	} else {
		if !flag {
			c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("user permission deny")))
			return
		} else {
			var d []*openapi.DomainInfo
			err := c.BindJSON(&d)
			if err != nil {
				c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
				return
			}

			res, err := h.service.ModDomain(d, name, domains)
			if err != nil {
				c.JSON(http.StatusOK, frame.NewOpenApiInternalError(res, err))
				return
			}
			c.JSON(http.StatusOK, frame.NewResponse(res))
		}
	}
}

// DelDomain 删除域名
func (h *handler) DelDomain(c *gin.Context) {
	name := c.GetHeader("x-client-id")
	if name == "" {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("get user failed")))
		return
	}
	flag, domains, permError := h.service.IsUserValidate(name)
	if permError != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(permError))
		return
	} else {
		if !flag {
			c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("user permission deny")))
			return
		} else {
			d := openapi.DelDomainParam{}
			err := c.BindJSON(&d)
			if err != nil {
				c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
				return
			}

			res, err := h.service.DelDomain(&d, name, domains)
			if err != nil {
				c.JSON(http.StatusOK, frame.NewOpenApiInternalError(res, err))
				return
			}
			c.JSON(http.StatusOK, frame.NewResponse(res))
		}
	}
}
