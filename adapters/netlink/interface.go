package netlink

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/pterm/pterm"
	"github.com/vishvananda/netlink"
)

func (s *networkService) InterfaceCreate(name string, mac string) error {
	pterm.DefaultSpinner.Info(fmt.Sprintf("ℹ️  Creating network interface: %s with mac %s\n", name, mac))

	link := &netlink.Tuntap{
		LinkAttrs: netlink.LinkAttrs{
			Name: name,
		},
		Mode: netlink.TUNTAP_MODE_TAP,
	}

	if mac != "" {
		addr, err := net.ParseMAC(mac)
		if err != nil {
			return fmt.Errorf("parsing mac address: %w", err)
		}
		link.Attrs().HardwareAddr = addr
	}

	if err := netlink.LinkAdd(link); err != nil {
		return fmt.Errorf("creating network interface %s: %w", link.Attrs().Name, err)
	}

	linkDetails, err := netlink.LinkByName(link.Attrs().Name)
	if err != nil {
		return fmt.Errorf("getting interface %s: %w", link.Attrs().Name, err)
	}

	if err := netlink.LinkSetUp(linkDetails); err != nil {
		return fmt.Errorf("setting network interface to UP %s: %w", linkDetails.Attrs().Name, err)
	}

	return nil
}

func (s *networkService) InterfaceDelete(name string) error {
	pterm.DefaultSpinner.Info(fmt.Sprintf("ℹ️  Deleting network interface: %s\n", name))

	return deleteLink(name, "interface")
}

func (s *networkService) InterfaceExists(name string) (bool, error) {
	pterm.DefaultSpinner.Info(fmt.Sprintf("ℹ️  Checking if network interface exists: %s\n", name))

	return linkExists(name, "interface")
}

func (s *networkService) NewInterfaceName(prefix string) (string, error) {
	slog.Debug("Generating new network interface name")

	index := 0
	breakGlassIndex := 1000 //TODO: make this configurable?
	for {
		name := fmt.Sprintf("%s%d", prefix, index)
		exists, err := s.InterfaceExists(name)
		if err != nil {
			return "", fmt.Errorf("checking if interface %s exists: %w", name, err)
		}
		if !exists {
			return name, nil
		}
		index++

		if index >= breakGlassIndex {
			return "", fmt.Errorf("failed to generate interface name, hit limit 0f %d", breakGlassIndex)
		}
	}
}
