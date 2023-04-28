package proxy

type Proxy interface {
	AsProxy() Proxy
	Run() error
}
