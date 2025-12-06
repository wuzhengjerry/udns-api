// Package docs 文档模块
package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://test.udns.woa.com",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/flow/status_notify": {
            "get": {
                "description": "QFlow状态推送",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "状态推送"
                ],
                "summary": "状态推送",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "推送ID",
                        "name": "_qflow_task_id",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/flow/task": {
            "post": {
                "description": "传递参数新增task",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "流程管理"
                ],
                "summary": "新增流程任务",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "新增task参数",
                        "name": "object",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/flow.StartTaskParams"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/flow/task/{task_id}": {
            "get": {
                "description": "根据taskID查询task详情",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "流程管理"
                ],
                "summary": "task详情",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "task ID,流程ID",
                        "name": "task_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/flow/task/{task_id}/steps/{step_id}/submit": {
            "put": {
                "description": "根据step ID提交步骤，通常在保存字段后操作",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "流程管理"
                ],
                "summary": "提交步骤",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "step ID,步骤ID",
                        "name": "step_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "task ID,任务ID",
                        "name": "task_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/flow/task/{task_id}/stop": {
            "put": {
                "description": "根据taskID中止步骤，只有creator和管理员可以操作",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "流程管理"
                ],
                "summary": "中止流程",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "step ID,步骤ID",
                        "name": "task_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/flow/tasks": {
            "post": {
                "description": "查询所有流程列表，需要管理员权限",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "流程列表"
                ],
                "summary": "流程列表",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "查询task参数",
                        "name": "object",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/flow.QueryTaskParams"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/flow/tasks/my_application": {
            "post": {
                "description": "查询我发起所有流程，支持各种参数",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "流程列表"
                ],
                "summary": "我的申请",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "查询task参数",
                        "name": "object",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/flow.QueryTaskParams"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/flow/tasks/my_approve": {
            "post": {
                "description": "查询由我审批的流程，包括当前待我审批，以及我审批完成的流程",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "流程列表"
                ],
                "summary": "我的审批",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "查询task参数",
                        "name": "object",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/flow.QueryTaskParams"
                        }
                    },
                    {
                        "type": "string",
                        "description": "当done=true查询已经审批完成的流程，否则查询当前待我审批的列表",
                        "name": "done",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/flow/tasks/steps/{step_id}/field": {
            "put": {
                "description": "根据步骤ID修改字段值",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "流程管理"
                ],
                "summary": "保存字段",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "参数：{field_id: field_value}",
                        "name": "object",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object"
                        }
                    },
                    {
                        "type": "string",
                        "description": "step ID,步骤ID",
                        "name": "step_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/flows": {
            "get": {
                "description": "查询所有流程类型，为发起流程查询流程提供参数",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "流程列表"
                ],
                "summary": "流程类型列表",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/role": {
            "get": {
                "description": "查询角色列表",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "权限管理"
                ],
                "summary": "角色列表",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "page size",
                        "name": "ps",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "page number",
                        "name": "pn",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "基础域关键字，支持模糊匹配，可为空",
                        "name": "name",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/role/{id}": {
            "put": {
                "description": "根据ID编辑具体角色对应用户",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "权限管理"
                ],
                "summary": "编辑角色",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "编辑线路参数",
                        "name": "object",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/role.Role"
                        }
                    },
                    {
                        "type": "string",
                        "description": "角色 ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/route": {
            "get": {
                "description": "根据名称匹配查询线路列表",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "线路管理"
                ],
                "summary": "线路列表",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "page size",
                        "name": "ps",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "page number",
                        "name": "pn",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "基础域关键字，支持模糊匹配，可为空",
                        "name": "name",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "传递参数新增线路",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "线路管理"
                ],
                "summary": "新增线路",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "新增线路参数",
                        "name": "object",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/route.Route"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/route/{id}": {
            "put": {
                "description": "根据ID编辑具体线路",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "线路管理"
                ],
                "summary": "编辑线路",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "编辑线路参数",
                        "name": "object",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/route.Route"
                        }
                    },
                    {
                        "type": "string",
                        "description": "线路 ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            },
            "delete": {
                "description": "根据ID删除具体线路",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "线路管理"
                ],
                "summary": "删除线路",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "线路 ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/zone": {
            "get": {
                "description": "根据名称匹配查询二级zone列表",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "二级zone管理"
                ],
                "summary": "二级zone列表",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "page size",
                        "name": "ps",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "page number",
                        "name": "pn",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "基础域关键字，支持模糊匹配，可为空",
                        "name": "name",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "传递参数新增二级zone",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "二级zone管理"
                ],
                "summary": "新增二级zone",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "新增二级zone参数",
                        "name": "object",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/zone.Zone"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        },
        "/zone/{id}": {
            "put": {
                "description": "根据ID编辑具体二级zone",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "二级zone管理"
                ],
                "summary": "编辑二级zone",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "编辑二级zone参数",
                        "name": "object",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/zone.Zone"
                        }
                    },
                    {
                        "type": "string",
                        "description": "二级zone ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            },
            "delete": {
                "description": "根据ID删除具体二级zone",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "二级zone管理"
                ],
                "summary": "删除二级zone",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "StaffName",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "二级zone ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/frame.AuthResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "flow.Basic": {
            "type": "object",
            "properties": {
                "child_item": {
                    "type": "string"
                },
                "task_name": {
                    "type": "string"
                }
            }
        },
        "flow.QueryTaskParams": {
            "type": "object",
            "properties": {
                "create_time_end": {
                    "type": "string"
                },
                "create_time_start": {
                    "type": "string"
                },
                "creator": {
                    "type": "string"
                },
                "doing_step_owner": {
                    "type": "string"
                },
                "field_input_value": {
                    "type": "string"
                },
                "finish_time_end": {
                    "type": "string"
                },
                "finish_time_start": {
                    "type": "string"
                },
                "flow_id": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "name": {
                    "type": "string"
                },
                "page": {
                    "type": "integer"
                },
                "page_size": {
                    "type": "integer"
                },
                "status": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "step_owner": {
                    "type": "string"
                }
            }
        },
        "flow.StartTaskParams": {
            "type": "object",
            "properties": {
                "basic": {
                    "$ref": "#/definitions/flow.Basic"
                },
                "echo_fields": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "echo_step_owners": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "flow_id": {
                    "type": "integer"
                },
                "flow_type": {
                    "type": "string"
                }
            }
        },
        "frame.AuthResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {
                    "type": "object"
                },
                "msg": {
                    "type": "string"
                }
            }
        },
        "role.Role": {
            "type": "object",
            "properties": {
                "created_time": {
                    "description": "创建时间",
                    "type": "integer"
                },
                "deleted": {
                    "description": "是否删除",
                    "type": "boolean"
                },
                "deleted_time": {
                    "description": "删除时间",
                    "type": "integer"
                },
                "description": {
                    "description": "描述信息",
                    "type": "string"
                },
                "id": {
                    "description": "ID",
                    "type": "string"
                },
                "name": {
                    "description": "线路名称",
                    "type": "string"
                },
                "updated_time": {
                    "description": "恢复时间",
                    "type": "integer"
                },
                "users": {
                    "description": "关联用户，分号相隔",
                    "type": "string"
                }
            }
        },
        "route.Route": {
            "type": "object",
            "properties": {
                "created_time": {
                    "description": "创建时间",
                    "type": "integer"
                },
                "deleted": {
                    "description": "是否删除",
                    "type": "boolean"
                },
                "deleted_time": {
                    "description": "删除时间",
                    "type": "integer"
                },
                "description": {
                    "description": "描述信息",
                    "type": "string"
                },
                "enabled": {
                    "description": "是否启用",
                    "type": "boolean"
                },
                "id": {
                    "description": "ID",
                    "type": "string"
                },
                "name": {
                    "description": "线路名称",
                    "type": "string"
                },
                "operator": {
                    "description": "负责人",
                    "type": "string"
                },
                "route_id": {
                    "description": "线路ID",
                    "type": "string"
                },
                "updated_time": {
                    "description": "恢复时间",
                    "type": "integer"
                }
            }
        },
        "zone.Zone": {
            "type": "object",
            "properties": {
                "created_time": {
                    "description": "触发时间",
                    "type": "integer"
                },
                "deleted": {
                    "description": "是否删除",
                    "type": "boolean"
                },
                "deleted_time": {
                    "description": "删除时间",
                    "type": "integer"
                },
                "description": {
                    "description": "描述信息",
                    "type": "string"
                },
                "enabled": {
                    "description": "是否启用",
                    "type": "boolean"
                },
                "id": {
                    "description": "ID",
                    "type": "string"
                },
                "name": {
                    "description": "二级zone域名",
                    "type": "string"
                },
                "operator": {
                    "description": "负责人",
                    "type": "string"
                },
                "updated_time": {
                    "description": "恢复时间",
                    "type": "integer"
                },
                "zone_id": {
                    "description": "二级zone id",
                    "type": "integer"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "test.udns.woa.com",
	BasePath:    "/api/v1/",
	Schemes:     []string{},
	Title:       "UDNS-WEB api服务",
	Description: "UDNS WEB 后台服务，主要提供域名管理、线路管理、后台管理等接口功能。",
}

type s struct{}

// ReadDoc 读文档
func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

// init 初始化
func init() {
	swag.Register(swag.Name, &s{})
}
