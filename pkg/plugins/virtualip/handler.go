package virtualip

import (
	"encoding/json"
	"fmt"
	"github.com/QQGoblin/veteran/pkg/config"
	"github.com/QQGoblin/veteran/pkg/plugins/virtualip/network"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
)

var (
	Name = "virtual_ip"
)

type VirtualIP struct {
	id      raft.ServerID
	handler network.Configurator
}

type virtualIPConfig struct {
	IFace   string `json:"iface"`
	Address string `json:"address"`
}

func (p *VirtualIP) Setup(veteranC *config.VeteranConfig) error {

	var err error

	p.id = raft.ServerID(veteranC.ID)

	tempConfig := struct {
		C virtualIPConfig `json:"virtual_ip"`
	}{}

	if err = json.Unmarshal(veteranC.Raw, &tempConfig); err != nil {
		return err
	}

	if tempConfig.C.IFace == "" || tempConfig.C.Address == "" {
		return fmt.Errorf("parameter configuration is incorrect")
	}

	p.handler, err = network.NewAliasConfigurator(tempConfig.C.Address, tempConfig.C.IFace)
	return err

}

func (p *VirtualIP) Handler(observation *raft.Observation) error {

	_, leader := observation.Raft.LeaderWithID()

	isSetVirtualIP, err := p.handler.IsSet()
	if err != nil {
		return err
	}

	// 1. 未设置 VIP, 连接 leader 失败
	// 2. 未设置 VIP, 当前不是 leader
	if !isSetVirtualIP && (leader == "" || leader != p.id) {
		return nil
	}

	// 3. 未设置 VIP, 当前是 leader
	if !isSetVirtualIP && leader == p.id {
		log.WithField("name", Name).Info("[Plugin] add virtual ip")
		if err = p.handler.AddIP(); err != nil {
			return err
		}
	}

	// 4. 设置 VIP, 当前是 leader
	if isSetVirtualIP && leader == p.id {
		return nil
	}

	// 5. 设置 VIP, 连接 leader 失败
	// 6. 设置 VIP, 当前不是 leader
	if isSetVirtualIP && (leader == "" || leader != p.id) {
		log.WithField("name", Name).Info("[Plugin] delete virtual ip")
		if err = p.handler.DeleteIP(); err != nil {
			return err
		}
	}

	return nil

}

func (p *VirtualIP) Shutdown() error {

	if err := p.handler.DeleteIP(); err != nil {
		return err
	}
	return nil
}

func (p *VirtualIP) Name() string { return Name }

func (p *VirtualIP) Filter(_ *raft.Observation) bool { return true }
