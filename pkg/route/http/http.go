// Package http 路由操作
package http

import (
	"errors"

	"github.com/wuzhengjerry/udns-api/frame/router"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/route"
)

var (
	api = &handler{}
)

// 服务集
type handler struct {
	service route.Service
}

// Registry 注册HTTP服务路由
func (h *handler) Registry(n *router.Nengine) {
	r := n.Group("/api/v1")
	n.Handle(r, "POST", "route", route.BasePermGroup, false, h.CreateRoute)
	n.Handle(r, "DELETE", "route/:id", route.BasePermGroup, false, h.DeleteRoute)
	n.Handle(r, "PUT", "route/:id", route.BasePermGroup, false, h.UpdateRoute)
	n.Handle(r, "GET", "route", route.BasePermGroup, false, h.QueryRoutes)
}

// Config 配置
func (h *handler) Config() error {
	if pkg.Route == nil {
		return errors.New("dependence application module Route is nil")
	}

	h.service = pkg.Route
	return nil
}

// init 初始化
func init() {
	pkg.RegistryHTTPV1("route", api)
}
