package resolver

import "context"

type HostnameIPMapping map[string]string

type Filter struct {
	Name    string
	Labels  []string
	Mapping map[string]string
}

type Inspector interface {
	GetContainerMapping(ctx context.Context, filter Filter) (HostnameIPMapping, error)
}
