package pkg

import (
	"fmt"

	"github.com/wuzhengjerry/udns-api/pkg/domain"
	"github.com/wuzhengjerry/udns-api/pkg/flow"
	"github.com/wuzhengjerry/udns-api/pkg/openapi"
	"github.com/wuzhengjerry/udns-api/pkg/role"
	"github.com/wuzhengjerry/udns-api/pkg/route"
	"github.com/wuzhengjerry/udns-api/pkg/webuser"
	"github.com/wuzhengjerry/udns-api/pkg/zone"
)

var (
	// Zone 二级域服务
	Zone zone.Service
	// Route 线路服务
	Route route.Service
	// Role 角色服务
	Role role.Service
	// Flow 审批服务
	Flow flow.Service
	// Domain 域名服务
	Domain domain.Service
	// WebUser 用户
	WebUser webuser.Service
	// OpenApi 对外接口服务
	OpenApi openapi.Service
)

var (
	servers       []Service
	successLoaded []string
)

// addService 加载服务
func addService(name string, svr Service) {
	servers = append(servers, svr)
	successLoaded = append(successLoaded, name)
}

// Service 注册上的服务必须实现的方法
type Service interface {
	Config() error
}

// RegistryService 服务实例注册
func RegistryService(name string, svr Service) {
	switch value := svr.(type) {
	case zone.Service:
		if Zone != nil {
			registryError(name)
		}
		Zone = value
		addService(name, svr)
	case route.Service:
		if Route != nil {
			registryError(name)
		}
		Route = value
		addService(name, svr)
	case role.Service:
		if Role != nil {
			registryError(name)
		}
		Role = value
		addService(name, svr)
	case flow.Service:
		if Flow != nil {
			registryError(name)
		}
		Flow = value
		addService(name, svr)
	case domain.Service:
		if Domain != nil {
			registryError(name)
		}
		Domain = value
		addService(name, svr)
	case webuser.Service:
		if WebUser != nil {
			registryError(name)
		}
		WebUser = value
		addService(name, svr)
	case openapi.Service:
		if OpenApi != nil {
			registryError(name)
		}
		OpenApi = value
		addService(name, svr)
	default:
		fmt.Println(value)
		panic(fmt.Sprintf("unknown module type %s", name))
	}
}

// registryError 服务注册失败
func registryError(name string) {
	panic("module " + name + " has registried")
}

// InitService 初始化所有服务
func InitService() error {
	for _, s := range servers {
		if err := s.Config(); err != nil {
			return err
		}
	}
	return nil
}
