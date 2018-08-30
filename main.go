package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opencopilot/packet-ip-sidecar/ip"
)

func main() {
	err := ip.AddDummy()
	if err != nil {
		log.Println(err)
	}
	log.Println("added packet0 interface")

	quit := make(chan bool, 1)
	go ip.EnsureIPs(quit)

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	<-gracefulStop
	log.Println("received stop signal, shutting down")
	quit <- true

	err = ip.RemoveDummy()
	if err != nil {
		log.Println(err)
	}
	log.Println("removed packet0 interface")

	time.Sleep(1 * time.Second)
	os.Exit(0)
}
