package stun

import (
	"fmt"
	"net"

	"github.com/pion/stun/v3"
	// "net/netip"
)

type SomeName struct {
	conn net.PacketConn
}

const stunSize int = 200
const udpSize int = 1400

func Process(conn net.PacketConn) error {
	for {

		pb := make([]byte, udpSize+stunSize)
		_, addr, err := conn.ReadFrom(pb)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("addr:", addr.String())
		msg := stun.New()
		err = stun.Decode(pb, msg)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("message", msg)

	}
}

func Listen(address string) (net.PacketConn, error) {

	udp_conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(address), Port: 3481})
	if err != nil {
		return nil, err
	}
	fmt.Println(udp_conn.LocalAddr())
	return udp_conn, nil
}
