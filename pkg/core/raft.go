package core

import (
	"fmt"
	"github.com/hashicorp/raft"
	"net"
	"time"
)

const MemberOper = time.Second * 3

type Manager struct {
	id     string
	bind   string
	logger Logger
	peers  map[string]string
	fsm    FSM
	Raft   *raft.Raft
}

type ClusterState struct {
	Members  []raft.Server `json:"Members"`
	ID       string        `json:"ID"`
	LeaderID string        `json:"LeaderID"`
	Status   string        `json:"Status"`
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

func (mgr *Manager) Status() (*ClusterState, error) {

	_, leaderID := mgr.Raft.LeaderWithID()

	future := mgr.Raft.GetConfiguration()

	if err := future.Error(); err != nil {
		return nil, err
	}

	return &ClusterState{
		Members:  future.Configuration().Servers,
		ID:       mgr.id,
		LeaderID: string(leaderID),
		Status:   mgr.Raft.State().String(),
	}, nil

}

func (mgr *Manager) AddMember(memberID, address string) error {

	if err := mgr.leaderCheck(); err != nil {
		return err
	}

	cstate, err := mgr.Status()
	if err != nil {
		return err
	}

	for _, member := range cstate.Members {
		if raft.ServerID(memberID) == member.ID {
			return nil
		}
		if raft.ServerAddress(address) == member.Address {
			return fmt.Errorf("%s conflict with %s", memberID, member.ID)
		}
	}

	// TODO: 单节点添加第二个 Member 时， 需要先启动被添加节点否者会导致服务不可用
	return mgr.Raft.AddVoter(raft.ServerID(memberID), raft.ServerAddress(address), 0, MemberOper).Error()
}

func (mgr *Manager) DelMember(memberID string) error {

	if err := mgr.leaderCheck(); err != nil {
		return err
	}

	return mgr.Raft.RemoveServer(raft.ServerID(memberID), 0, MemberOper).Error()
}

func (mgr *Manager) leaderCheck() error {

	_, leaderID := mgr.Raft.LeaderWithID()

	if raft.ServerID(mgr.id) != leaderID {
		return fmt.Errorf("only add member on leader")
	}

	return nil
}

func (mgr *Manager) inCluster(memberID raft.ServerID) bool {

	cstate, err := mgr.Status()
	if err != nil {
		return false
	}

	for _, member := range cstate.Members {
		if member.ID == memberID {
			return true
		}
	}
	return false
}
