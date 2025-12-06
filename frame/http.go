package frame

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/wuzhengjerry/udns-api/conf"
	_ "github.com/wuzhengjerry/udns-api/docs" //
	"github.com/wuzhengjerry/udns-api/frame/router"
	log "github.com/wuzhengjerry/udns-api/logs"
	"github.com/wuzhengjerry/udns-api/pkg"
)

const (
	USER   = "user"   // USER 来自web端用户的请求
	MODULE = "module" // MODULE 来自系统内部模块的请求
	CLIENT = "client" // CLIENT 来自外部api调用的请求
)

// NewHTTPService 构建函数
func NewHTTPService() *HTTPService {
	if conf.C().App.Mode != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(log.LoggerToFile(), gin.Recovery())
	engine.GET("api/v1/swagger/*any", ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "UDNS_DISABLE_SWAGGER"))

	ends := make(map[string]*router.Endpoint)
	n := &router.Nengine{
		Engine:     engine,
		Endpoints:  ends,
		PathPrefix: "", // 这里在初始router时，直接设定全局prefix
	}
	// 增加允许跨域中间件
	n.Use(n.Cors)

	return &HTTPService{
		e: n,
		l: log.C(),
		c: conf.C(),
	}
}

// HTTPService http服务
type HTTPService struct {
	e *router.Nengine
	l *logrus.Logger
	c *conf.Config
}

// Start 启动server
func (s *HTTPService) Start() error {
	if err := pkg.InitV1HTTPAPI(s.c.App.Name, s.e); err != nil {
		return err
	}
	log.C().WithFields(logrus.Fields{}).Info("[START] start http server...")

	return s.e.Engine.Run(s.c.App.Host + ":" + s.c.App.Port)
}

// Stop 停止server
func (s *HTTPService) Stop() error {
	s.l.Info("[STOP] start graceful shutdown")
	os.Exit(123)

	return nil
}
