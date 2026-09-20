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

func Process(conn *net.UDPConn) error {
	for {

		pb := make([]byte, udpSize+stunSize)
		_, addr, err := conn.ReadFromUDPAddrPort(pb)
		port := addr.Port()
		bin_ipv4 := addr.Addr().As4()
		fmt.Printf("bin %x\n", bin_ipv4)
		var ipv4addr uint32
		ipv4addr = uint32(bin_ipv4[3]) << 24
		ipv4addr |= (uint32(bin_ipv4[2]) << 16)
		ipv4addr |= (uint32(bin_ipv4[1]) << 8)
		ipv4addr |= uint32(bin_ipv4[0])
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
		xor_mapped_port := port ^ uint16(magic_cookie>>16)
		xor_mapped_address := ipv4addr ^ magic_cookie
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
		// func xormapping(){

		// }
		fmt.Printf("xor: %x %x\n", xor_mapped_address, xor_mapped_port)
		var send = stun.New()
		xmap := stun.XORMappedAddress{IP: net.ParseIP(addr.Addr().String())}
		xmap.AddTo(send)
		fmt.Printf("send %v", send.Attributes)
		_, err = conn.WriteToUDPAddrPort(append([]byte{}, byte(xor_mapped_address>>24), byte(xor_mapped_address<<8>>24),
			byte(xor_mapped_address<<16>>24), byte(xor_mapped_address<<24>>24), byte(','), byte(' '),
			byte(xor_mapped_port>>8), byte(xor_mapped_port<<8>>8)), addr)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("message %v \n transaction %x", msg, msg.TransactionID)
		fmt.Printf("\nrexor: %x.%x.%x.%x  %x\n", (xor_mapped_address^magic_cookie)>>24, (xor_mapped_address^magic_cookie)<<8>>24, (xor_mapped_address^magic_cookie)<<16>>24, (xor_mapped_address^magic_cookie)<<24>>24, xor_mapped_port^uint16(magic_cookie>>16))

	}
}

func Listen(address string) (*net.UDPConn, error) {

	udp_conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(address), Port: 3481})
	if err != nil {
		return nil, err
	}
	fmt.Println(udp_conn.LocalAddr())
	return udp_conn, nil
}
