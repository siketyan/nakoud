package discovery

import (
	"context"

	"github.com/siketyan/nakoud/pkg/proto/ip"
)

type Discoverer interface {
	Discover(ctx context.Context, fqdn string) (*ip.Address, error)
}
