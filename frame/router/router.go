// Package router 路由模块
package router

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Endpoint 终端结构
type Endpoint struct {
	Path    string
	Method  string
	PermKey string
	IsAuth  bool
}

// Nengine 引擎
type Nengine struct {
	Engine     *gin.Engine
	Endpoints  map[string]*Endpoint
	PathPrefix string
}

// Group 对于gin RouterGroup的封装
func (n *Nengine) Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	relativePath = n.PathPrefix + "/" + relativePath
	return n.Engine.Group(relativePath, handlers...)
}

// Handle 对于gin Handle函数的封装
// 在此处对endpoint添加了isAuth，permType等权限标示，为后面权限校验函数提供基础
func (n *Nengine) Handle(r *gin.RouterGroup, method, relativePath, permKey string, isAuth bool,
	handlers ...gin.HandlerFunc) gin.IRoutes {
	endpoint := &Endpoint{
		Path:    r.BasePath() + relativePath,
		Method:  method,
		PermKey: permKey,
		IsAuth:  isAuth,
	}
	// 使用method_path作为endpoint的唯一标示key，例如"GET_/role/:id"
	endpointKey := endpoint.Method + "_" + endpoint.Path
	n.Endpoints[endpointKey] = endpoint

	return r.Handle(method, relativePath, handlers...)
}

// Use 封装gin的Use中间件方法
func (n *Nengine) Use(middleware ...gin.HandlerFunc) {
	n.Engine.Use(middleware...)
}

// Cors 跨站检查
func (n *Nengine) Cors(c *gin.Context) {
	method := c.Request.Method
	origin := c.Request.Header.Get("Origin") //请求头部
	if origin != "" {
		//接收客户端发送的origin （重要！）
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		//服务器支持的所有跨域请求的方法
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
		//允许跨域设置可以返回其他子段，可以自定义字段
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, "+
			"Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, "+
			"X-Requested-With,BusinessID, Grant-Type, StaffName")
		// 允许浏览器（客户端）可以解析的头部 （重要）
		c.Header("Access-Control-Expose-Headers", "Content-Type, Content-Length, "+
			"Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, "+
			"X-Requested-With,BusinessID, Grant-Type, StaffName")
		//设置缓存时间
		c.Header("Access-Control-Max-Age", "172800")
		//允许客户端传递校验信息比如 cookie (重要)
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	//允许类型校验
	if method == "OPTIONS" {
		c.JSON(http.StatusOK, "ok!")
	}

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic info is: %v", err)
		}
	}()

	c.Next()
}
