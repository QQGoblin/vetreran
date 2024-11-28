package metadata

import (
	"bytes"
	"encoding/json"
	"github.com/QQGoblin/veteran/pkg/config"
	"github.com/hashicorp/raft"
	"os"
	"path"
)

var (
	Name = "metadata"
)

const (
	output = "metadata.json"
)

type MemberStatus struct {
	Offline bool               `json:"offline,omitempty"` //TODO: 在非 leader 节点, Online 信息并不准确
	Address raft.ServerAddress `json:"address"`
	ID      raft.ServerID      `json:"id"`
}
type Metadata struct {
	Leader  raft.ServerID  `json:"leader"` // 节点离线时，leader 为空
	Members []MemberStatus `json:"members"`
	Output  string         `json:"-"`
}

func (p *Metadata) Setup(config *config.VeteranConfig) error {
	p.Output = path.Join(config.Store, output)
	return nil
}

func (p *Metadata) Handler(observation *raft.Observation) error {

	_, p.Leader = observation.Raft.LeaderWithID()
	future := observation.Raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return err
	}

	c := future.Configuration()
	p.Members = make([]MemberStatus, len(c.Servers))

	for i, server := range c.Servers {

		Offline := false
		if failedHeartbeat, isOK := observation.Data.(raft.FailedHeartbeatObservation); isOK {
			Offline = failedHeartbeat.PeerID == server.ID
		}

		p.Members[i] = MemberStatus{
			Offline: Offline,
			Address: server.Address,
			ID:      server.ID,
		}
	}

	o, err := json.Marshal(p)
	if err != nil {
		return err
	}
	var str bytes.Buffer
	_ = json.Indent(&str, o, "", "    ")

	return os.WriteFile(p.Output, str.Bytes(), 0600)

}

func (p *Metadata) Shutdown() error {

	if err := os.Remove(p.Output); err != nil && os.IsNotExist(err) {
		return err
	}
	return nil
}

func (p *Metadata) Name() string { return Name }

func (p *Metadata) Filter(_ *raft.Observation) bool { return true }
