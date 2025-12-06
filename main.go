package main

import (
	"fmt"

	"github.com/wuzhengjerry/udns-api/cmd"
)

// @title UDNS-WEB api服务
// @version 1.0
// @description  UDNS WEB 后台服务，主要提供域名管理、线路管理、后台管理等接口功能。
// @termsOfService http://localhost:8060
// @BasePath /api/v1/
// @host localhost
func main() {
	cmd.Execute()
	fmt.Println("udns service started")
}
