package aws

import (
	"context"

	"go-micro.dev/v4/registry"
)

type servicesKey struct{}


// Services is an option that preloads service data
func Services(s map[string][]*registry.Service) registry.Option {
	return func(o *registry.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, servicesKey{}, s)
	}
}
