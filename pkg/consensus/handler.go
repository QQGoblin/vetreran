package consensus

import (
	"fmt"
	"github.com/hashicorp/raft"
)

func (m *Manager) Status() (*ClusterState, error) {

	if m.Raft == nil {
		return nil, fmt.Errorf("raft is not init")
	}

	_, leaderID := m.Raft.LeaderWithID()

	future := m.Raft.GetConfiguration()

	if err := future.Error(); err != nil {
		return nil, err
	}

	return &ClusterState{
		Members:  future.Configuration().Servers,
		ID:       m.id,
		LeaderID: string(leaderID),
		Status:   m.Raft.State().String(),
	}, nil

}

func (m *Manager) AddMember(memberID, address string, nonVoter bool) error {

	if err := m.leaderCheck(); err != nil {
		return err
	}

	cstate, err := m.Status()
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
	m.lock.Lock()
	defer m.lock.Unlock()

	if nonVoter {
		return m.Raft.AddNonvoter(raft.ServerID(memberID), raft.ServerAddress(address), 0, memberOperTimeout).Error()
	}

	return m.Raft.AddVoter(raft.ServerID(memberID), raft.ServerAddress(address), 0, memberOperTimeout).Error()
}

func (m *Manager) DelMember(memberID string) error {

	if err := m.leaderCheck(); err != nil {
		return err
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	return m.Raft.RemoveServer(raft.ServerID(memberID), 0, memberOperTimeout).Error()
}

func (m *Manager) leaderCheck() error {

	if m.Raft == nil {
		return fmt.Errorf("raft is not init")
	}

	_, leaderID := m.Raft.LeaderWithID()

	if raft.ServerID(m.id) != leaderID {
		return fmt.Errorf("only add member on leader")
	}

	return nil
}
