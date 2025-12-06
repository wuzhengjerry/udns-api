package route

// Service 角色服务
type Service interface {
	InsertRoute(r *Route) error
	DeleteRoute(id string) (err error)
	UpdateRoute(r *Route) (err error)
	DescribeRoute(id string) (*Route, error)
	QueryRoutes(name string, ps, pn int64) (count int64, rs []*Route, err error)
	IsRouteExisted(id string) (ok bool, err error)
}
