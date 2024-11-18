package pkg

import (
	"github.com/QQGoblin/veteran/pkg/core"
	log "github.com/sirupsen/logrus"
	"time"
)

type Veteran struct {
	config      *VeteranConfig
	floatingMGR *FloatingMGR
	core        *core.Manager
	stop        chan bool
	finished    chan bool
}

const (
	DataPath = "/opt/veteran"
)

func NewVeteran(config *VeteranConfig) (*Veteran, error) {

	core, err := core.NewManager(config.ID, config.Bind, config.Peers)
	if err != nil {
		return nil, err
	}

	floatingMGR, err := NewFloatingMGR(config.Floating)
	if err != nil {
		return nil, err
	}

	return &Veteran{
		config:      config,
		floatingMGR: floatingMGR,
		core:        core,
	}, nil
}

func (v *Veteran) Start() error {

	// 初始化 Raft
	if err := v.core.InitRaft(); err != nil {
		return err
	}

	v.stop = make(chan bool, 1)
	v.finished = make(chan bool, 1)

	go v.loop()

	log.Info("Started")

	return nil
}

func (v *Veteran) Stop() {
	close(v.stop)

	<-v.finished

	log.Info("Stopped")
}

func (v *Veteran) loop() {

	ticker := time.NewTicker(time.Second)
	isLeader := false

	v.floatingMGR.deleteIP()

	for {
		select {
		case leader := <-v.core.Raft.LeaderCh():
			if leader {
				isLeader = true
				log.Info("Leading")
				v.floatingMGR.addIP()
			} else {
				isLeader = false
				log.Info("Following")
				v.floatingMGR.deleteIP()
			}

		case <-ticker.C:
			if isLeader {
				result, err := v.floatingMGR.isSet()
				if err != nil {
					log.WithError(err).WithField("ip", v.config.Floating.Address).Error("Could not check ip")
				}
				if result == false {
					log.Error("Lost IP")
					v.floatingMGR.addIP()
				}
			}

		case <-v.stop:
			log.Info("Stopping")
			if isLeader {
				v.floatingMGR.deleteIP()
			}
			close(v.finished)
			return
		}
	}
}
