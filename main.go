package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/packethost/packetmetadata/packetmetadata"

	"github.com/vishvananda/netlink"
)

func addIP(link netlink.Link, addr string) error {
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

func removeIP(link netlink.Link, addr string) error {
	a, err := netlink.ParseAddr(addr)
	if err != nil {
		return err
	}
	err = netlink.AddrDel(link, a)
	if err != nil {
		return err
	}
	fmt.Println("removing IP: " + addr)
	return nil
}

func ensureIPs(quit chan bool) {
	iterator, err := packetmetadata.Watch()
	if err != nil {
		log.Println(err)
	}

	lo, err := netlink.LinkByName("lo")
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

			existingAddrs, err := netlink.AddrList(lo, 4)
			if err != nil {
				log.Println(err)
			}

			incomingAddrs := res.Metadata.Instance.Network

			// iterate over IPs that should be added
			for _, addr := range incomingAddrs {

				ipBlock := addr.Address + "/" + strconv.Itoa(addr.Cidr)
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
					err := addIP(lo, ipBlock)
					if err != nil {
						log.Println(err)
					}
				}
			}

			// iterate over IPs that are already added
			for _, addr := range existingAddrs {

				shouldKeep := false
				for _, incomingAddr := range incomingAddrs {
					ipBlock := incomingAddr.Address + "/" + strconv.Itoa(incomingAddr.Cidr)
					incomingAddr, err := netlink.ParseAddr(ipBlock)
					if err != nil {
						log.Println(err)
					}
					if addr.Equal(*incomingAddr) {
						shouldKeep = true
					}
				}

				if !shouldKeep { // if ip exists locally, but should not
					removeIP(lo, addr.String())
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}
}

func main() {
	quit := make(chan bool, 1)
	go ensureIPs(quit)

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	<-gracefulStop
	log.Println("received stop signal, shutting down")
	quit <- true
	time.Sleep(1 * time.Second)
	os.Exit(0)
}
