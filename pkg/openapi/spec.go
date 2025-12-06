package openapi

import "go.mongodb.org/mongo-driver/bson"

// Service OpenApi信息
type Service interface {
	QueryDomain(q *QueryParam) (count int64, rs []*DomainInfo, err error)
	AddDomain(d []*DomainInfo, user string, p []string) (res interface{}, err error)
	ModDomain(d []*DomainInfo, user string, p []string) (res interface{}, err error)
	DelDomain(d *DelDomainParam, user string, p []string) (res interface{}, err error)
	IsUserValidate(name string) (flag bool, domains []string, err error)
	IsParamValidate(p *DomainInfo) (flag bool, err error)
	ValidateQueryParam(p *QueryParam) (res bson.M, err error)
}
