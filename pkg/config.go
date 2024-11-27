package pkg

import (
	"encoding/json"
	"os"
)

type VeteranConfig struct {
	ID        string            `json:"id"`
	Listen    string            `json:"listen"`
	Store     string            `json:"store"`
	InitPeers map[string]string `json:"peers"`
	Floating  *Floating         `json:"floating"`
	RaftLog   RaftLogConfig     `json:"raft_log"`
}

type VirtualType string

const (
	AliasType          VirtualType = "alias"
	MACVlanType        VirtualType = "macvlan"
	defaultRaftLogName             = "/var/log/veteran/raft.log"
)

type Floating struct {
	IFace   string      `json:"iface"`
	Type    VirtualType `json:"type"`
	Address string      `json:"address"`
}

type RaftLogConfig struct {
	Output string `json:"output"`
	Enable bool   `json:"enable"`
	Level  string `json:"level"`
}

func LoadConfig(filepath string) (*VeteranConfig, error) {

	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	c := &VeteranConfig{
		RaftLog: RaftLogConfig{
			Output: defaultRaftLogName,
			Enable: false,
			Level:  "info",
		},
	}

	if err = json.Unmarshal(b, c); err != nil {
		return nil, err
	}

	if c.ID == "" {
		c.ID, _ = os.Hostname()
	}

	return c, nil
}
