package core

import (
	"github.com/hashicorp/raft"
	"net"
	"time"
)

type Manager struct {
	id     string
	bind   string
	logger Logger
	peers  map[string]string
	fsm    FSM
	Raft   *raft.Raft
}

func NewManager(id, bind string, peers map[string]string) (*Manager, error) {

	return &Manager{
		id:     id,
		bind:   bind,
		logger: Logger{},
		peers:  peers,
		fsm:    FSM{},
	}, nil

}

func (mgr *Manager) InitRaft() error {
	// Create configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(mgr.id)
	config.LogOutput = mgr.logger

	// Initialize communication
	address, err := net.ResolveTCPAddr("tcp", mgr.bind)
	if err != nil {
		return err
	}

	// Create transport
	transport, err := raft.NewTCPTransport(mgr.bind, address, 3, 10*time.Second, mgr.logger)
	if err != nil {
		return err
	}

	// Create Raft structures
	snapshots := raft.NewInmemSnapshotStore()
	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	// Cluster configuration
	configuration := raft.Configuration{}

	for id, ip := range mgr.peers {
		configuration.Servers = append(configuration.Servers, raft.Server{ID: raft.ServerID(id), Address: raft.ServerAddress(ip)})
	}

	// Bootstrap cluster
	if err = raft.BootstrapCluster(config, logStore, stableStore, snapshots, transport, configuration); err != nil {
		return err
	}

	// Create RAFT instance
	raftServer, err := raft.NewRaft(config, mgr.fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return err
	}

	mgr.Raft = raftServer

	return nil
}
