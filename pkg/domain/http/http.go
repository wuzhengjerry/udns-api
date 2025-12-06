// Package http 域名操作
package http

import (
	"errors"

	"github.com/wuzhengjerry/udns-api/frame/router"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/domain"
)

var (
	api = &handler{}
)

// handler 服务集
type handler struct {
	service domain.Service
}

// Registry 注册HTTP服务路由
func (h *handler) Registry(n *router.Nengine) {
	r := n.Group("/api/v1")
	n.Handle(r, "POST", "sync_it_api", domain.BasePermGroup, false, h.SyncITApi)
	n.Handle(r, "GET", "sync_it_log", domain.BasePermGroup, false, h.QuerySyncApiLogs)
	n.Handle(r, "POST", "get_config", domain.BasePermGroup, false, h.GetDomainConfig)
	n.Handle(r, "POST", "add_config", domain.BasePermGroup, false, h.AddDomainConfig)
	n.Handle(r, "POST", "modify_config", domain.BasePermGroup, false, h.ModifyDomainConfig)
	n.Handle(r, "POST", "delete_config", domain.BasePermGroup, false, h.DeleteDomainConfig)
	n.Handle(r, "POST", "domainlog", domain.BasePermGroup, false, h.CreateDomainLog)
	n.Handle(r, "DELETE", "domainlog/:id", domain.BasePermGroup, false, h.DeleteDomainLog)
	n.Handle(r, "GET", "domainlog", domain.BasePermGroup, false, h.QueryDomainLogs)
	n.Handle(r, "POST", "domain", domain.BasePermGroup, false, h.CreateDomain)
	n.Handle(r, "DELETE", "domain/:name", domain.BasePermGroup, false, h.DeleteDomain)
	n.Handle(r, "PUT", "domain/:name", domain.BasePermGroup, false, h.UpdateDomain)
	n.Handle(r, "GET", "domain", domain.BasePermGroup, false, h.QueryDomains)
	n.Handle(r, "GET", "domain/:name", domain.BasePermGroup, false, h.DescribeDomain)
	n.Handle(r, "GET", "get_all_staff_fullname", domain.BasePermGroup, false, h.GetAllStaffFullName)
	n.Handle(r, "GET", "get_it_data", domain.BasePermGroup, false, h.GetITData)
	n.Handle(r, "GET", "get_staff_info", domain.BasePermGroup, false, h.GetStaffInfo)
	n.Handle(r, "GET", "get_business_tree", domain.BasePermGroup, false, h.GetBusinessTree)
	n.Handle(r, "GET", "is_domain_existed", domain.BasePermGroup, false, h.IsDomainsExisted)
	n.Handle(r, "GET", "is_oadomain_existed", domain.BasePermGroup, false, h.QueryITDomainsExisted)
}

// Config 配置
func (h *handler) Config() error {
	if pkg.Domain == nil {
		return errors.New("dependence application module Domain is nil")
	}

	h.service = pkg.Domain
	return nil
}

// init 初始化
func init() {
	pkg.RegistryHTTPV1("domain", api)
}
