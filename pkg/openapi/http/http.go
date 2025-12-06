// Package http 后端调用模块
package http

import (
	"errors"

	"github.com/wuzhengjerry/udns-api/frame/router"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/openapi"
)

var (
	api = &handler{}
)

// handler 服务集
type handler struct {
	service openapi.Service
}

// Registry 注册HTTP服务路由
func (h *handler) Registry(n *router.Nengine) {
	r := n.Group("/openapi/v1")
	n.Handle(r, "POST", "query_domain", openapi.BasePermGroup, false, h.QueryDomain)
	n.Handle(r, "POST", "add_domain", openapi.BasePermGroup, false, h.AddDomain)
	n.Handle(r, "POST", "mod_domain", openapi.BasePermGroup, false, h.ModDomain)
	n.Handle(r, "POST", "del_domain", openapi.BasePermGroup, false, h.DelDomain)
}

// Config 配置
func (h *handler) Config() error {
	if pkg.OpenApi == nil {
		return errors.New("dependence application module OpenApi is nil")
	}

	h.service = pkg.OpenApi
	return nil
}

// init 初始化
func init() {
	pkg.RegistryHTTPV1("openapi", api)
}
