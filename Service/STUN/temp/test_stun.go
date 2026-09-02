package main

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/stun/v3"
)

func main() {
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.ParseIP("localhost"), Port: 3481})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", conn.RemoteAddr())
	stc, err := stun.NewClient(conn)
	if err != nil {
		panic(err)
	}
	tc := time.NewTicker(time.Second * 2)
	defer tc.Stop()
	for _ = range tc.C {

		if err := stc.Start(stun.MustBuild(stun.BindingRequest, stun.TransactionID), func(e stun.Event) { fmt.Println(e) }); err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("Persisting")
}
