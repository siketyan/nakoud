package discovery

import (
	"context"

	"github.com/siketyan/nakoud/pkg/upstream"
)

type Discoverer interface {
	Discover(ctx context.Context, host []string) (upstream.Client, error)
}
