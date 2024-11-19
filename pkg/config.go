package pkg

import (
	"encoding/json"
	"fmt"
	"os"
)

type VeteranConfig struct {
	ID       string            `json:"id"`
	Bind     string            `json:"bind"`
	Listen   string            `json:"listen"`
	Peers    map[string]string `json:"peers"`
	Floating *Floating         `json:"floating"`
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

	if c.Peers == nil {
		return nil, fmt.Errorf("peers is empty")
	}

	if c.ID == "" {
		c.ID, _ = os.Hostname()
	}

	address, isOK := c.Peers[c.ID]
	if !isOK {
		return nil, fmt.Errorf("current nodeid<%s> is not found", c.ID)
	}

	c.Bind = address

	return c, nil
}
