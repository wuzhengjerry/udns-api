package frame

// AuthResponse 鉴权返回包
type AuthResponse struct {
	RetCode int32       `json:"code"`
	Data    interface{} `json:"data"`
	Msg     string      `json:"msg"`
}

// PageData 分页
type PageData struct {
	Count int         `json:"count"`
	Data  interface{} `json:"data"`
	Pn    int         `json:"pn"`
	Ps    int         `json:"ps"`
}

// OpenApiPageData openapi请求
type OpenApiPageData struct {
	Count  int         `json:"count"`
	Data   interface{} `json:"data"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
}

// FlowResponse 审批流响应包
type FlowResponse struct {
	Ret  int32       `json:"ret"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// NewFlowResponse 审批流请求包
func NewFlowResponse(data interface{}) *FlowResponse {
	return &FlowResponse{
		Ret:  0,
		Data: data,
		Msg:  "",
	}
}

// NewFlowFailResponse  审批流请求失败的响应包
func NewFlowFailResponse(data interface{}) *FlowResponse {
	return &FlowResponse{
		Ret:  2,
		Data: data,
		Msg:  "task operate failed",
	}
}

// NewResponse  审批流请求响应包
func NewResponse(data interface{}) *AuthResponse {
	return &AuthResponse{
		RetCode: 0,
		Data:    data,
		Msg:     "Success",
	}
}

// NewPagedResponse 响应包
func NewPagedResponse(data interface{}, ps, pn, count int) *AuthResponse {
	pageData := PageData{
		Count: count,
		Pn:    pn,
		Ps:    ps,
		Data:  data,
	}
	return &AuthResponse{
		RetCode: 0,
		Data:    pageData,
		Msg:     "",
	}
}

// NewOpenApiResponse openapi响应包
func NewOpenApiResponse(data interface{}, limit, offset, count int) *AuthResponse {
	pageData := OpenApiPageData{
		Count:  count,
		Offset: offset,
		Limit:  limit,
		Data:   data,
	}
	return &AuthResponse{
		RetCode: 0,
		Data:    pageData,
		Msg:     "",
	}
}

// NewNotFound 资源不存在响应包
func NewNotFound(err error) *AuthResponse {
	return &AuthResponse{
		RetCode: 1000404,
		Data:    nil,
		Msg:     err.Error(),
	}
}

// NewBadRequest 不合理请求响应包
func NewBadRequest(err error) *AuthResponse {
	return &AuthResponse{
		RetCode: 1000400,
		Data:    nil,
		Msg:     err.Error(),
	}
}

// NewPermissionDenied 权限非法响应
func NewPermissionDenied(err error) *AuthResponse {
	return &AuthResponse{
		RetCode: 1000403,
		Data:    nil,
		Msg:     err.Error(),
	}
}

// NewInternalError 内部错误响应
func NewInternalError(err error) *AuthResponse {
	return &AuthResponse{
		RetCode: 1000500,
		Data:    nil,
		Msg:     err.Error(),
	}
}

// NewOpenApiInternalError openapi内部错误响应
func NewOpenApiInternalError(data interface{}, err error) *AuthResponse {
	return &AuthResponse{
		RetCode: 1000500,
		Data:    data,
		Msg:     err.Error(),
	}
}
