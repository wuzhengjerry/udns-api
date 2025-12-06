// Package http 用户管理
package http

import (
	"errors"

	"github.com/wuzhengjerry/udns-api/frame/router"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/webuser"
)

var (
	api = &handler{}
)

// handler 服务集
type handler struct {
	service webuser.Service
}

// Registry 注册HTTP服务路由
func (h *handler) Registry(n *router.Nengine) {
	r := n.Group("/api/v1")
	n.Handle(r, "POST", "user", webuser.BasePermGroup, false, h.CreateUser)
	n.Handle(r, "DELETE", "user/:name", webuser.BasePermGroup, false, h.DeleteUser)
	n.Handle(r, "PUT", "user/:name", webuser.BasePermGroup, false, h.UpdateUser)
	n.Handle(r, "GET", "user", webuser.BasePermGroup, false, h.QueryUsers)
	n.Handle(r, "GET", "user/:name", webuser.BasePermGroup, false, h.DescribeUser)
}

// Config 配置
func (h *handler) Config() error {
	if pkg.WebUser == nil {
		return errors.New("dependence application module WebUser is nil")
	}

	h.service = pkg.WebUser
	return nil
}

// init 初始化
func init() {
	pkg.RegistryHTTPV1("webuser", api)
}
