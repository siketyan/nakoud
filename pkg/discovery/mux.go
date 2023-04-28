package discovery

import (
	"context"
	"fmt"

	"github.com/siketyan/nakoud/pkg/proto/ip"
)

type Mux struct {
	discoverers []Discoverer
}

func NewMux() *Mux {
	return &Mux{
		discoverers: make([]Discoverer, 0),
	}
}

func (m *Mux) Add(d Discoverer) {
	m.discoverers = append(m.discoverers, d)
}

func (m *Mux) With(d Discoverer) *Mux {
	m.Add(d)

	return m
}

func (m *Mux) Discover(ctx context.Context, fqdn string) (*ip.Address, error) {
	for _, d := range m.discoverers {
		address, err := d.Discover(ctx, fqdn)
		if err != nil {
			return nil, fmt.Errorf("error occurred while discovering upstream: %w", err)
		}

		if address != nil {
			return address, nil
		}
	}

	return nil, nil
}
