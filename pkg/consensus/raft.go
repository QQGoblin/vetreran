package consensus

import (
	"fmt"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"io"
	"net"
	"os"
	"path"
	"sync"
	"time"
)

const (
	memberOperTimeout = time.Second * 3
	db                = "veteran.db"
)

type Manager struct {
	id        string
	logger    io.Writer
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
		logger:    io.Discard,
		storePath: storePath,
		initPeers: initPeers,
		fsm:       FSM{},
	}, nil

}

func (m *Manager) InitRaft() error {
	// Create configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(m.id)
	config.LogOutput = m.logger

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

	snapshots := raft.NewInmemSnapshotStore()
	existing, err := raft.HasExistingState(boltDB, boltDB, snapshots)
	if err != nil {
		return err
	}

	// 初始化 raft 集群
	if !existing {
		return m.newCluster(config, boltDB)
	}

	// 启动 raft 集群
	temp := raft.DefaultConfig()
	temp.LocalID = raft.ServerID(m.id)
	temp.LogOutput = m.logger
	configuration, err := raft.GetConfiguration(temp, m.fsm, boltDB, boltDB, raft.NewInmemSnapshotStore(), &raft.InmemTransport{})
	if err != nil {
		return err
	}

	var bind string
	for _, server := range configuration.Servers {
		if server.ID == raft.ServerID(m.id) {
			bind = string(server.Address)
		}
	}

	if bind == "" {
		return fmt.Errorf("this node is not found in raft.db")
	}

	address, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return err
	}

	transport, err := raft.NewTCPTransport(bind, address, 3, 10*time.Second, m.logger)
	if err != nil {
		return err
	}

	m.Raft, err = raft.NewRaft(config, m.fsm, boltDB, boltDB, snapshots, transport)
	return err
}

func (m *Manager) newCluster(config *raft.Config, store *raftboltdb.BoltStore) error {

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

	transport, err := raft.NewTCPTransport(bind, address, 3, 10*time.Second, m.logger)
	if err != nil {
		return err
	}

	snapshots := raft.NewInmemSnapshotStore()

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
