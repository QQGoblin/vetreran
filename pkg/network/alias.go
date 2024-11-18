package network

import (
	"github.com/j-keck/arping"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

type AliasConfigurator struct {
	link    netlink.Link
	address *netlink.Addr
	iface   string
}

func NewAliasConfigurator(address, iface string) (result AliasConfigurator, error error) {
	addr, error := netlink.ParseAddr(address)
	if error != nil {
		error = errors.Wrapf(error, "could not parse address '%s'", address)

		return
	}

	result = AliasConfigurator{iface: iface, address: addr}

	result.link, error = netlink.LinkByName(iface)
	if error != nil {
		error = errors.Wrapf(error, "could not get link for interface '%s'", iface)

		return
	}

	return
}

func (configurator AliasConfigurator) AddIP() error {
	result, error := configurator.IsSet()
	if error != nil {
		return errors.Wrap(error, "ip check in AddIP failed")
	}

	// Already set
	if result {
		return nil
	}

	if error = netlink.AddrAdd(configurator.link, configurator.address); error != nil {
		return errors.Wrap(error, "could not add ip")
	}

	if error = arping.GratuitousArpOverIfaceByName(configurator.address.IP, configurator.iface); error != nil {
		return errors.Wrap(error, "gratuitous arp failed")
	}

	return nil
}

func (configurator AliasConfigurator) DeleteIP() error {
	result, error := configurator.IsSet()
	if error != nil {
		return errors.Wrap(error, "ip check in DeleteIP failed")
	}

	// Nothing to delete
	if !result {
		return nil
	}

	if error = netlink.AddrDel(configurator.link, configurator.address); error != nil {
		return errors.Wrap(error, "could not delete ip")
	}

	return nil
}

func (configurator AliasConfigurator) IsSet() (result bool, error error) {
	var addresses []netlink.Addr

	addresses, error = netlink.AddrList(configurator.link, 0)
	if error != nil {
		error = errors.Wrap(error, "could not list addresses")

		return
	}

	for _, address := range addresses {
		if address.Equal(*configurator.address) {
			return true, nil
		}
	}

	return false, nil
}
