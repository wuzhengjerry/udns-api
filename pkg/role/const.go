package role

import "time"

const (
	HalfHourTTL = 30 * time.Minute

	BasePermGroup = "base_perm_group"

	SuperAdminName RoleName = "super_admin"
	FlowAdminName  RoleName = "flow_admin"
)
