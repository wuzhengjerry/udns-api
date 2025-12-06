// Package http 角色管理模块
package http

import (
	"errors"

	"github.com/wuzhengjerry/udns-api/frame/router"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/role"
)

var (
	api = &handler{}
)

// handler 服务集
type handler struct {
	service role.Service
}

// Registry 注册HTTP服务路由
func (h *handler) Registry(n *router.Nengine) {
	r := n.Group("/api/v1")
	//n.Handle(r, "POST", "Role", role.BasePermGroup, false, h.CreateRole)
	//n.Handle(r, "DELETE", "Role/:id", role.BasePermGroup, false, h.DeleteRole)
	n.Handle(r, "PUT", "role/:id", role.BasePermGroup, false, h.UpdateRole)
	n.Handle(r, "GET", "role", role.BasePermGroup, false, h.QueryRoles)
}

// Config 配置
func (h *handler) Config() error {
	if pkg.Role == nil {
		return errors.New("dependence application module Role is nil")
	}

	h.service = pkg.Role
	return nil
}

// init 初始化
func init() {
	pkg.RegistryHTTPV1("role", api)
}
