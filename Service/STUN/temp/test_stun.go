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
	go func() {
		p := make([]byte, 124)
		for {
			_, addr, err := conn.ReadFrom(p)
			if err == nil {
				fmt.Print("stun?", stun.IsMessage(p))
				msg := stun.New()
				err := stun.Decode(p, msg)
				if err != nil {
					fmt.Print(err, "Error:decoding")
					continue
				}
				fmt.Println("Stn:", msg)
				fmt.Println(addr.String(), "\n")
				// fmt.Printf("%x", p)
				fmt.Println("Stn:", msg.Attributes[0])
				continue
			}
			fmt.Println("exiting", err)
			return
		}
	}()
	defer tc.Stop()
	for _ = range tc.C {
		fmt.Println("\n-----Sending----")
		if err := stc.Start(stun.MustBuild(stun.BindingRequest, stun.TransactionID), func(e stun.Event) { fmt.Println(e) }); err != nil {
			fmt.Println(err)
			continue
		}

	}
}
