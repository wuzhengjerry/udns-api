// Package qflow 审批结构体
package qflow

import (
	"github.com/sirupsen/logrus"
	log "github.com/wuzhengjerry/udns-api/logs"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/domain"
	"github.com/wuzhengjerry/udns-api/pkg/flow"
)

var (
	Service = &service{}
)

// service 服务集
type service struct {
	l *logrus.Logger
	d domain.Service
}

// Config 配置
func (s *service) Config() error {

	s.l = log.C()
	s.d = pkg.Domain
	return nil
}

// init 初始化
func init() {
	var _ flow.Service = Service
	pkg.RegistryService("flow", Service)
}
