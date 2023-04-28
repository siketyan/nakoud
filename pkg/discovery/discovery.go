package discovery

import (
	"github.com/siketyan/nakoud/pkg/upstream"
)

type Discoverer interface {
	Discover(host []string) (upstream.Client, error)
}
