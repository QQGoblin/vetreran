package plugins

import (
	"context"
	"github.com/QQGoblin/veteran/pkg/config"
	"github.com/QQGoblin/veteran/pkg/plugins/metadata"
	"github.com/QQGoblin/veteran/pkg/plugins/virtualip"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
)

var (
	Plugins = make(map[string]Plugin)
)

type Plugin interface {
	Filter(*raft.Observation) bool
	Handler(*raft.Observation) error
	Setup(config *config.VeteranConfig) error
	Shutdown() error
	Name() string
}

func init() {
	Register(metadata.Name, &metadata.Metadata{})
	Register(virtualip.Name, &virtualip.VirtualIP{})
}

func Register(name string, plugin Plugin) {
	if name == "" {
		log.Fatalln("Plugin must have a name")
	}
	if _, ok := Plugins[name]; !ok {
		Plugins[name] = plugin
	} else {
		log.Fatalln("Plugin name conflict")
	}
}

func StartPlugin(ctx context.Context, input chan raft.Observation, p Plugin) {

	go func() {
		log.WithField("name", p.Name()).Info("Start plugin success")
		for {
			select {
			case observation := <-input:
				if err := p.Handler(&observation); err != nil {
					log.WithError(err).WithField("name", p.Name()).Error("Plugin failed to run")
				}
			case <-ctx.Done():
				if err := p.Shutdown(); err != nil {
					log.WithError(err).WithField("name", p.Name()).Error("Plugin failed to shutdown")
				}
				return
			}
		}
	}()
}
