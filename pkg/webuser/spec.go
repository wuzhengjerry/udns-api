package webuser

// Service 角色服务
type Service interface {
	InsertUser(z *WebUser) error
	DeleteUser(name string) (err error)
	UpdateUser(z *WebUser) (err error)
	DescribeUser(name string) (r *WebUser, err error)
	QueryUsers(name string, ps, pn int64) (count int64, zs []*WebUser, err error)
}
