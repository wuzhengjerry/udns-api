// Package all 初始化模块
package all

import (
	// 加载服务模块
	_ "github.com/wuzhengjerry/udns-api/pkg/domain/http"
	_ "github.com/wuzhengjerry/udns-api/pkg/domain/mongo"
	_ "github.com/wuzhengjerry/udns-api/pkg/flow/http"
	_ "github.com/wuzhengjerry/udns-api/pkg/flow/qflow"
	_ "github.com/wuzhengjerry/udns-api/pkg/openapi/http"
	_ "github.com/wuzhengjerry/udns-api/pkg/openapi/mongo"
	_ "github.com/wuzhengjerry/udns-api/pkg/role/http"
	_ "github.com/wuzhengjerry/udns-api/pkg/role/mongo"
	_ "github.com/wuzhengjerry/udns-api/pkg/route/http"
	_ "github.com/wuzhengjerry/udns-api/pkg/route/mongo"
	_ "github.com/wuzhengjerry/udns-api/pkg/webuser/http"
	_ "github.com/wuzhengjerry/udns-api/pkg/webuser/mongo"
	_ "github.com/wuzhengjerry/udns-api/pkg/zone/http"
	_ "github.com/wuzhengjerry/udns-api/pkg/zone/mongo"
)
