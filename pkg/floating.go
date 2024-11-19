package pkg

import (
	"fmt"
	"github.com/QQGoblin/veteran/pkg/network"
	log "github.com/sirupsen/logrus"
)

type FloatingMGR struct {
	config  *Floating
	handler network.Configurator
}

func NewFloatingMGR(c *Floating) (*FloatingMGR, error) {

	var (
		err     error
		handler network.Configurator
	)

	switch c.Type {
	case AliasType:
		handler, err = network.NewAliasConfigurator(c.Address, c.IFace)
	case MACVlanType:
	default:
		err = fmt.Errorf("FloatingMGR %s is not support", c.Type)
	}

	if err != nil {
		log.WithError(err).Error("Init FloatingMGR")
		return nil, err
	}

	return &FloatingMGR{
		config:  c,
		handler: handler,
	}, nil

}

func (manager *FloatingMGR) addIP() {
	if err := manager.handler.AddIP(); err != nil {
		log.WithError(err).Error("Could not set ip")
	} else {
		log.WithField("address", manager.config.Address).Info("Added IP")
	}
}

func (manager *FloatingMGR) deleteIP() {
	if err := manager.handler.DeleteIP(); err != nil {
		log.WithError(err).Error("Could not delete ip")
	} else {
		log.WithField("address", manager.config.Address).Info("Deleted IP")
	}
}

func (manager *FloatingMGR) isSet() (bool, error) {
	return manager.handler.IsSet()
}
