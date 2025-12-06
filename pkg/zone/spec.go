package zone

// Service 角色服务
type Service interface {
	InsertZone(z *Zone) error
	DeleteZone(id string) (err error)
	UpdateZone(z *Zone) (err error)
	QueryZones(name string, ps, pn int64) (count int64, zs []*Zone, err error)
	DescribeZone(name string) (d *Zone, err error)
}
