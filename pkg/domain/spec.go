package domain

// Service 域名信息
type Service interface {
	InsertDomain(user string, d *Domain) error
	DeleteDomain(user, name string) (err error)
	UpdateDomain(user string, d *Domain) (err error)
	DescribeDomain(name string) (d *Domain, err error)
	QueryDomains(manage, name, owner, operator, description, businessId string,
		ps, pn int64) (count int64, rs []*Domain, err error)
	InsertDomainLog(user string, d *DomainLog) error
	DeleteDomainLog(id string) (err error)
	DescribeDomainLog(id string) (r *DomainLog, err error)
	QueryDomainLogs(name string, ps, pn int64) (count int64, rs []*DomainLog, err error)
	GetAllStaffFullName() (res string, err error)
	GetStaffInfo(user string) (res map[string]interface{}, err error)
	GetBusinessTree() (res map[string]interface{}, err error)
	GetDomainConfig(params *QueryDomainParam) (res interface{}, err error)
	BatchGetDomainConfig(params *BatchQueryDomainParam) (res []*ModDomainParam, err error)
	AddDomainConfig(params *ModDomainParam) (res interface{}, err error)
	DeleteDomainConfig(user string, params *DelDomainParam) (res interface{}, err error)
	DeleteDomainConfigAction(params *DelDomainParam) (res interface{}, err error)
	ModifyDomainConfig(name string, params *ModDomainParam) (res interface{}, err error)
	ValidateParam(user, flowType string, data map[string]string) (flag bool, msg string, err error)
	DomainOperation(user, flowType string, data map[string]string) (result map[string]Message, err error)
	HasDomainPermission(name, user string) bool
	IsDomainsExisted(name string) (msg map[string]interface{}, err error)
	QueryITDomainsExisted(names, user string) (res Message, err error)
	SyncITApi(user string, params *ITParam) (res ITApiResp, err error)
	QuerySyncApiLogs(name, user, method, apiType, success string, ps, pn int64) (
		count int64, rs []*SyncApiLog, err error)
	SyncUdnsApiLog(user, params, domainName, apiType, method, success, msg string) (err error)
	GetDomainDefaultConfig(domainName string) (res RouteConfig, flag bool, err error)
	SplitStr(ss string) (res string)
	GenITDomainData(user string, optType string, new *RouteConfig, old *RouteConfig) (res ITDomainInfo)
	GenITApiData(user string, d []ITDomainInfo) (res ITParam)
	RequestRITApi(domainName string) (res ITRApiResp, err error)
	ModDomainConfig(params *ModDomainParam) (res interface{}, err error)
	IsOADomainExisted(user, name string) (res Message, err error)
}
