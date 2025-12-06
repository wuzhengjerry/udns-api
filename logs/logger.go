package log

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/conf"
)

// Logger 写日志
func Logger() *logrus.Logger {
	logConfig := conf.C().Log
	logFileName := conf.C().App.Name + ".log"
	//日志文件
	fileName := path.Join(logConfig.PathDir, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			fmt.Println("aaa", err.Error())
		}
	}
	//实例化
	logger := logrus.New()
	//设置输出
	writer, err := rotatelogs.New(
		fileName+".%Y%m%d",
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithRotationCount(60),
		rotatelogs.WithRotationTime(time.Hour*24),
	)

	if err != nil {
		fmt.Println("init logger err", err)
	}

	logger.SetOutput(writer)
	logger.SetFormatter(&logrus.JSONFormatter{})
	//设置日志级别
	logger.SetLevel(logrus.DebugLevel)
	//设置日志格式
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return logger
}

// LoggerToFile 日志写入文件
func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		// 处理请求
		c.Next()
		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime).String()
		// 请求方式
		reqMethod := c.Request.Method
		// 请求路由
		reqUri := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.ClientIP()
		// 请求主体，user,module,client
		user := c.GetHeader("StaffName")

		// 健康检查日志不记录
		if reqUri == "/openapi/v1" && reqMethod == "GET" {
			return
		}

		if reqUri == "/api/v1" && reqMethod == "GET" {
			return
		}

		if reqUri == "/" && reqMethod == "GET" {
			return
		}

		logField := logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUri,
			"user":         user,
		}
		openApiUser := c.GetHeader("x-client-id")
		if openApiUser != "" {
			logField["openapi"] = openApiUser
		}
		l.WithFields(logField).Info()
	}
}
