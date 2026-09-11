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
		data := pb[:]
		messagetype := data[:2]
		// messagetype[0]=messagetype[0]<<2;
		var messagetypefull uint16
		fmt.Println("For")
		for i := range messagetype {
			fmt.Printf("%x", messagetype[i])
		}
		fmt.Println("\n---")
		messagetypefull = uint16(messagetype[0] << 10)

		messagetypefull |= uint16(messagetype[1])
		messagelength := data[2:4]
		var magic_cookie uint32
		mcb := data[4:8]
		magic_cookie = uint32(mcb[0]) << 24
		magic_cookie |= (uint32(mcb[1]) << 16)
		magic_cookie |= (uint32(mcb[2]) << 8)
		magic_cookie |= (uint32(mcb[3]))
		trcb := data[8:20]
		var transactionID1 uint64
		var transactionID2 uint32
		transactionID1 = uint64(trcb[0]) << 56
		for i := 2; i <= 8; i++ {

			transactionID1 |= (uint64(trcb[i-1]) << (64 - (8 * i)))
			// transactionID1 |= (uint64(trcb[1]) << 40)
			// transactionID1 |= (uint64(trcb[1]) << 32)
			// transactionID1 |= (uint64(trcb[1]) << 32)
		}
		transactionID2 = (uint32(trcb[8]) << 24)
		transactionID2 |= (uint32(trcb[9]) << (32 - 16))
		transactionID2 |= (uint32(trcb[10]) << (32 - 24))
		transactionID2 |= uint32(trcb[11])

		fmt.Printf("Message type %x \n Message length %x\n  magic_cookie %x\n transactionID:%x%x", messagetypefull, messagelength, magic_cookie, transactionID1, transactionID2)
		fmt.Println("\naddr:", addr.String())
		msg := stun.New()
		err = stun.Decode(pb, msg)
		if err != nil {
			fmt.Println(err)
			return err
		}
		// tp:=stun.NewType(stun.MethodBinding,stun.BindingSuccess.Class)
		// m2:=stun.NewWithOptions(stun.WithStrict())
		// (tp)
		_, err = conn.WriteTo([]byte("message recieved"), addr)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("message %v \n transaction %x", msg, msg.TransactionID)

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
