package role

// Service 角色服务
type Service interface {
	InsertRole(r *Role) error
	DeleteRole(id string) (err error)
	UpdateRole(r *Role) (err error)
	DescribeRole(name RoleName) (r *Role, err error)
	QueryRoles(name string, ps, pn int64) (count int64, rs []*Role, err error)
	HasRolePermission(roleName RoleName, user string) bool
}
