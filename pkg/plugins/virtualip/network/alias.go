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

func NewAliasConfigurator(address, iface string) (result AliasConfigurator, err error) {
	addr, err := netlink.ParseAddr(address)
	if err != nil {
		err = errors.Wrapf(err, "could not parse address '%s'", address)

		return
	}

	result = AliasConfigurator{iface: iface, address: addr}

	result.link, err = netlink.LinkByName(iface)
	if err != nil {
		err = errors.Wrapf(err, "could not get link for interface '%s'", iface)

		return
	}

	return
}

func (configurator AliasConfigurator) AddIP() error {
	result, err := configurator.IsSet()
	if err != nil {
		return errors.Wrap(err, "ip check in AddIP failed")
	}

	// Already set
	if result {
		return nil
	}

	if err = netlink.AddrAdd(configurator.link, configurator.address); err != nil {
		return errors.Wrap(err, "could not add ip")
	}

	if err = arping.GratuitousArpOverIfaceByName(configurator.address.IP, configurator.iface); err != nil {
		return errors.Wrap(err, "gratuitous arp failed")
	}

	return nil
}

func (configurator AliasConfigurator) DeleteIP() error {
	result, err := configurator.IsSet()
	if err != nil {
		return errors.Wrap(err, "ip check in DeleteIP failed")
	}

	// Nothing to delete
	if !result {
		return nil
	}

	if err = netlink.AddrDel(configurator.link, configurator.address); err != nil {
		return errors.Wrap(err, "could not delete ip")
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
