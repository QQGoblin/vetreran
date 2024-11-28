package network

type Configurator interface {
	AddIP() error
	DeleteIP() error
	IsSet() (bool, error)
}
