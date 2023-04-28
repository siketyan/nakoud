package proxy

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type Mux struct {
	proxies []Proxy
}

func NewMux() *Mux {
	return &Mux{
		proxies: make([]Proxy, 0),
	}
}

func (m *Mux) Add(p Proxy) {
	m.proxies = append(m.proxies, p)
}

func (m *Mux) With(p Proxy) *Mux {
	m.Add(p)

	return m
}

func (m *Mux) AsProxy() Proxy {
	return m
}

func (m *Mux) Run() error {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(m.proxies))

	for _, p := range m.proxies {
		p := p
		go func() {
			if err := p.Run(); err != nil {
				log.Error().Err(err).Msg("A proxy server stopped with an error")
			}
		}()
	}

	waitGroup.Wait()

	return nil
}
