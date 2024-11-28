package pkg

import (
	"context"
	"fmt"
	"github.com/QQGoblin/veteran/pkg/config"
	"github.com/QQGoblin/veteran/pkg/consensus"
	logutils "github.com/QQGoblin/veteran/pkg/log"
	"github.com/QQGoblin/veteran/pkg/plugins"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type Veteran struct {
	config        *config.VeteranConfig
	core          *consensus.Manager
	srv           *http.Server
	pluginsCancel map[string]context.CancelFunc
}

func NewVeteran(c *config.VeteranConfig) (*Veteran, error) {

	core, err := consensus.NewManager(c.ID, c.InitPeers, c.Store)
	if err != nil {
		return nil, err
	}

	return &Veteran{
		config:        c,
		core:          core,
		pluginsCancel: make(map[string]context.CancelFunc),
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

	if err := v.initObserver(); err != nil {
		return err
	}

	go func() {
		if err := v.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Fatal("Start api server failure")
		}
	}()

	log.Info("Started")

	return nil
}

func (v *Veteran) Stop() {

	if v.srv != nil {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = v.srv.Shutdown(ctxWithTimeout)
		defer cancel()
	}

	for _, pluginCancel := range v.pluginsCancel {
		pluginCancel()
	}

	v.core.Shutdown()

	log.Info("Stopped")
}

func (v *Veteran) initObserver() error {

	if v.core.Raft == nil {
		return fmt.Errorf("raft is not initial")
	}

	for name, plugin := range plugins.Plugins {
		observationChan := make(chan raft.Observation)
		ctx, cancel := context.WithCancel(context.Background())
		v.pluginsCancel[name] = cancel
		if err := plugin.Setup(v.config); err != nil {
			log.WithError(err).WithField("name", name).Fatal("Setup plugin failure")
		}
		v.core.Raft.RegisterObserver(raft.NewObserver(observationChan, false, plugin.Filter))
		plugins.StartPlugin(ctx, observationChan, plugin)
	}

	return nil
}
