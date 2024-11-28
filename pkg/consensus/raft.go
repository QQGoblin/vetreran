package consensus

import (
	"fmt"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
	"path"
	"sync"
	"time"
)

const (
	memberOperTimeout = time.Second * 3
	applyOperTimeout  = time.Second * 3
	db                = "veteran.db"
	snapshotRetain    = 3
)

type Manager struct {
	id        string
	initPeers map[string]string
	storePath string
	fsm       FSM
	lock      sync.Mutex
	Raft      *raft.Raft
}

type ClusterState struct {
	Members  []raft.Server `json:"Members"`
	ID       string        `json:"ID"`
	LeaderID string        `json:"LeaderID"`
	Status   string        `json:"Status"`
}

func NewManager(id string, initPeers map[string]string, storePath string) (*Manager, error) {

	return &Manager{
		id:        id,
		storePath: storePath,
		initPeers: initPeers,
		fsm:       FSM{},
	}, nil

}

func (m *Manager) InitRaft(logger io.Writer, loglevel string) error {

	// 初始化配置
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(m.id)
	config.NoSnapshotRestoreOnStart = true // 只从快照文件初始集群配置信息
	config.LogOutput = logger
	config.LogLevel = loglevel

	// 初始化 snapshots 以及 DB
	if err := os.MkdirAll(m.storePath, 0755); err != nil {
		return err
	}

	boltDB, err := raftboltdb.New(raftboltdb.Options{
		Path: path.Join(m.storePath, db),
	})
	if err != nil {
		return fmt.Errorf("new bbolt store: %s", err)
	}

	snapshots, err := raft.NewFileSnapshotStore(m.storePath, snapshotRetain, logger)
	if err != nil {
		return fmt.Errorf("new snapshot store: %s", err)
	}

	existing, err := raft.HasExistingState(boltDB, boltDB, snapshots)
	if err != nil {
		return err
	}

	// 初始化 raft 集群
	if !existing {
		return m.newCluster(config, boltDB, snapshots)
	}

	// 启动 raft 集群
	transport, err := m.localTransport(boltDB, boltDB, snapshots, logger)
	if err != nil {
		return err
	}

	m.Raft, err = raft.NewRaft(config, m.fsm, boltDB, boltDB, snapshots, transport)
	return err
}

func (m *Manager) Shutdown() {
	shutdownFuture := m.Raft.Shutdown()
	if err := shutdownFuture.Error(); err != nil {
		log.WithError(err).Error("Stop raft failed")
	}
}

func (m *Manager) newCluster(config *raft.Config, store *raftboltdb.BoltStore, snapshots raft.SnapshotStore) error {

	// 初始化通信接口
	var bind string
	for id, address := range m.initPeers {
		if id == m.id {
			bind = address
		}
	}

	if bind == "" {
		return fmt.Errorf("this node is not found in init peers")
	}

	address, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return err
	}

	transport, err := raft.NewTCPTransport(bind, address, 3, 10*time.Second, config.LogOutput)
	if err != nil {
		return err
	}

	configuration := raft.Configuration{}
	for id, ip := range m.initPeers {
		configuration.Servers = append(configuration.Servers, raft.Server{ID: raft.ServerID(id), Address: raft.ServerAddress(ip)})
	}
	if err = raft.BootstrapCluster(config, store, store, snapshots, transport, configuration); err != nil {
		return err
	}

	// 启动服务
	m.Raft, err = raft.NewRaft(config, m.fsm, store, store, snapshots, transport)

	return err
}

func (m *Manager) localTransport(logStore raft.LogStore, stableStore raft.StableStore, snapshot raft.SnapshotStore, logger io.Writer) (raft.Transport, error) {

	temp := raft.DefaultConfig()
	temp.LocalID = raft.ServerID(m.id)
	temp.LogOutput = logger
	configuration, err := raft.GetConfiguration(temp, m.fsm, logStore, stableStore, snapshot, &raft.InmemTransport{})
	if err != nil {
		return nil, err
	}

	var bind string
	for _, server := range configuration.Servers {
		if server.ID == raft.ServerID(m.id) {
			bind = string(server.Address)
		}
	}

	if bind == "" {
		return nil, fmt.Errorf("this node is not found in raft.db")
	}

	if bind == "" {
		return nil, fmt.Errorf("this node is not found in raft.db")
	}

	address, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(bind, address, 3, 10*time.Second, logger)
	if err != nil {
		return nil, err
	}

	return transport, nil
}
