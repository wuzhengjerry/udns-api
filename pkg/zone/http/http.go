// Package http 地域管理
package http

import (
	"errors"

	"github.com/wuzhengjerry/udns-api/frame/router"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/zone"
)

var (
	api = &handler{}
)

// handler 服务集
type handler struct {
	service zone.Service
}

// Registry 注册HTTP服务路由
func (h *handler) Registry(n *router.Nengine) {
	r := n.Group("/api/v1")
	n.Handle(r, "POST", "zone", zone.BasePermGroup, false, h.CreateZone)
	n.Handle(r, "DELETE", "zone/:id", zone.BasePermGroup, false, h.DeleteZone)
	n.Handle(r, "PUT", "zone/:id", zone.BasePermGroup, false, h.UpdateZone)
	n.Handle(r, "GET", "zone", zone.BasePermGroup, false, h.QueryZones)
}

// Config 配置
func (h *handler) Config() error {
	if pkg.Zone == nil {
		return errors.New("dependence application module Zone is nil")
	}

	h.service = pkg.Zone
	return nil
}

// init 初始化
func init() {
	pkg.RegistryHTTPV1("zone", api)
}
