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
}

type VirtualType string

const (
	AliasType   VirtualType = "alias"
	MACVlanType VirtualType = "macvlan"
)

type Floating struct {
	IFace   string      `json:"iface"`
	Type    VirtualType `json:"type"`
	Address string      `json:"address"`
}

func LoadConfig(filepath string) (*VeteranConfig, error) {

	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	c := &VeteranConfig{}

	if err = json.Unmarshal(b, c); err != nil {
		return nil, err
	}

	if c.ID == "" {
		c.ID, _ = os.Hostname()
	}

	return c, nil
}
