package pkg

import (
	"context"
	"github.com/QQGoblin/veteran/pkg/consensus"
	logutils "github.com/QQGoblin/veteran/pkg/log"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type Veteran struct {
	config      *VeteranConfig
	floatingMGR *FloatingMGR
	core        *consensus.Manager
	srv         *http.Server
	stop        chan bool
	finished    chan bool
}

func NewVeteran(config *VeteranConfig) (*Veteran, error) {

	core, err := consensus.NewManager(config.ID, config.InitPeers, config.Store)
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

	// 初始化 API Server
	v.srv = v.apiServer()

	// 初始化 Raft
	logOutput := io.Discard
	if v.config.RaftLog.Enable {
		logOutput = logutils.RotateLogOutput(v.config.RaftLog.Output)
	}

	if err := v.core.InitRaft(logOutput, v.config.RaftLog.Level); err != nil {
		return err
	}

	v.stop = make(chan bool, 1)
	v.finished = make(chan bool, 1)

	go v.loop()

	go func() {
		if err := v.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Fatal("Start api server failure")
		}
	}()

	log.Info("Started")

	return nil
}

func (v *Veteran) Stop() {

	close(v.stop)

	<-v.finished

	if v.srv != nil {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = v.srv.Shutdown(ctxWithTimeout)
		defer cancel()
	}

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
