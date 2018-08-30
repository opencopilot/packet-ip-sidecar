package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/packethost/packetmetadata/packetmetadata"
	"github.com/packethost/packngo"
	"github.com/vishvananda/netlink"
)

var (
	doNotRemoveNets = []string{
		"fe80::3c50:1dff:fec5:4e5f/64",
		"169.254.0.0/16",
	}
	tag = "unknown"
)

// AddDummy creates a dummy interface called packet0
func AddDummy() error {
	err := netlink.LinkAdd(&netlink.Dummy{
		netlink.LinkAttrs{
			Name: "packet0",
		},
	})
	if err != nil {
		return err
	}

	dummy, err := netlink.LinkByName("packet0")
	if err != nil {
		return err
	}

	err = netlink.LinkSetUp(dummy)
	if err != nil {
		return err
	}
	return nil
}

// RemoveDummy removes a dummy interface called packet0
func RemoveDummy() error {
	dummy, err := netlink.LinkByName("packet0")
	if err != nil {
		return err
	}

	err = netlink.LinkDel(dummy)
	if err != nil {
		return err
	}
	return nil
}

// AddIP parses and adds an IP block to a network interface
func AddIP(link netlink.Link, addr string) error {
	a, err := netlink.ParseAddr(addr)
	if err != nil {
		return err
	}
	err = netlink.AddrReplace(link, a)
	if err != nil {
		return err
	}
	fmt.Println("adding IP: " + addr)
	return nil
}

// RemoveIP parses and remove an IP block from a network interface
func RemoveIP(link netlink.Link, addr string) error {
	ip, _, err := net.ParseCIDR(addr)
	if err != nil {
		return err
	}

	a, err := netlink.ParseAddr(addr)
	if err != nil {
		return err
	}

	// Do not remove the IP if it's in doNotRemoveNets
	for _, subnet := range doNotRemoveNets {
		_, ipnet, err := net.ParseCIDR(subnet)
		if err != nil {
			return err
		}
		if ipnet.Contains(ip) {
			return nil
		}
	}

	err = netlink.AddrDel(link, a)
	if err != nil {
		return err
	}
	fmt.Println("removing IP: " + addr)
	return nil
}

// EnsureIPs watches Packet metadata and ensures that any IP blocks added to an instance are added to the packet0 dummy interface
func EnsureIPs(quit chan bool) {
	iterator, err := packetmetadata.Watch()
	if err != nil {
		log.Println(err)
	}

	for {
		select {
		case <-quit:
			iterator.Close()
			break
		default:
			res, err := iterator.Next()
			if err != nil {
				log.Println(err)
			}

			packetIF, err := netlink.LinkByName("packet0")
			if err != nil {
				log.Println(err)
			}

			existingAddrs, err := netlink.AddrList(packetIF, netlink.FAMILY_ALL)
			if err != nil {
				log.Println(err)
			}

			incomingAddrs := make([]*packngo.IPAddressAssignment, 0)
			for _, incomingAddr := range res.Metadata.Instance.Network {
				if !incomingAddr.Management {
					incomingAddrs = append(incomingAddrs, incomingAddr)
				}
			}

			// iterate over IPs that should be added
			for _, addr := range incomingAddrs {

				ipBlock := addr.Address + "/" + strconv.Itoa(addr.CIDR)
				incomingAddr, err := netlink.ParseAddr(ipBlock)
				if err != nil {
					log.Println(err)
				}

				// look for IPs that are not currently added
				alreadyAdded := false
				for _, existingAddr := range existingAddrs {
					if existingAddr.Equal(*incomingAddr) {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded { // if ip does not exist locally, but should
					err := AddIP(packetIF, ipBlock)
					if err != nil {
						log.Println(err)
					}
				}
			}

			// iterate over IPs that are already added
			for _, addr := range existingAddrs {
				shouldKeep := false
				for _, incomingAddr := range incomingAddrs {
					ipBlock := incomingAddr.Address + "/" + strconv.Itoa(incomingAddr.CIDR)
					incomingAddr, err := netlink.ParseAddr(ipBlock)
					if err != nil {
						log.Println(err)
					}

					if addr.Equal(*incomingAddr) {
						shouldKeep = true
						break
					}
				}
				if !shouldKeep { // if ip exists locally, but should not
					err = RemoveIP(packetIF, strings.Split(addr.String(), " ")[0])
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}
}

func main() {
	err := AddDummy()
	if err != nil {
		log.Println(err)
	}

	quit := make(chan bool, 1)
	go EnsureIPs(quit)

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	<-gracefulStop
	log.Println("received stop signal, shutting down")
	quit <- true

	err = RemoveDummy()
	if err != nil {
		log.Println(err)
	}

	time.Sleep(1 * time.Second)
	os.Exit(0)
}
