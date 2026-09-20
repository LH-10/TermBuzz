package main

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/stun/v3"
)

func main() {
	conn, err := net.DialUDP("udp4", &net.UDPAddr{IP: net.ParseIP("localhost"), Port: 3409}, &net.UDPAddr{IP: net.ParseIP("localhost"), Port: 3481})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", conn.RemoteAddr())
	stc, err := stun.NewClient(conn)
	if err != nil {
		panic(err)
	}
	tc := time.NewTicker(time.Second * 2)
	p := make([]byte, 1024)
	go func() {
		for {
			_, addr, err := conn.ReadFrom(p)
			if err == nil {
				fmt.Println(addr.String(), "\n", string(p))
				continue
			}
			fmt.Println("exiting", err)
			return
		}
	}()
	defer tc.Stop()
	for {

		if err := stc.Start(stun.MustBuild(stun.BindingRequest, stun.TransactionID), func(e stun.Event) { fmt.Println(e) }); err != nil {
			fmt.Println(err)
			continue
		}

	}
}
