package config

import (
	"encoding/json"
	"os"
)

const (
	defaultRaftLogName = "/var/log/veteran/raft.log"
)

type VeteranConfig struct {
	ID        string            `json:"id"`
	Listen    string            `json:"listen"`
	Store     string            `json:"store"`
	InitPeers map[string]string `json:"initial_cluster"`
	RaftLog   RaftLogConfig     `json:"raft_log"`
	Raw       []byte
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
		Raw: b,
	}

	if err = json.Unmarshal(b, c); err != nil {
		return nil, err
	}

	if c.ID == "" {
		c.ID, _ = os.Hostname()
	}

	return c, nil
}
