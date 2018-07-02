package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/packethost/packngo/metadata"
	"github.com/vishvananda/netlink"
)

func getNetworkAddresses() ([]metadata.AddressInfo, error) {
	device, err := metadata.GetMetadata()
	if err != nil {
		return nil, err
	}

	addrs := make([]metadata.AddressInfo, 0)
	for _, addr := range device.Network.Addresses {
		if !addr.Management {
			addrs = append(addrs, addr)
		}
	}

	return addrs, nil
}

func ensureIPs() {
	addresses, err := getNetworkAddresses()
	if err != nil {
		fmt.Println(err)
	}

	lo, err := netlink.LinkByName("lo")
	if err != nil {
		fmt.Println(err)
	}

	for _, addr := range addresses {
		a, err := netlink.ParseAddr(addr.Address.String() + "/" + strconv.Itoa(addr.NetworkBits))
		if err != nil {
			fmt.Println(err)
		}
		err = netlink.AddrReplace(lo, a)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("ensuring IP: " + a.String())
	}
}

func main() {
	for {
		ensureIPs()
		time.Sleep(5 * time.Second)
	}
}
