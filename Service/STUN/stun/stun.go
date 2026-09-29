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

var magic_cookie uint32 = 0x2112A442

type Method uint16

const (
	MethodBinding Method = 0x01
)

type MessageClass uint16

const (
	Request        MessageClass = 0b00
	Indication     MessageClass = 0b01
	SuccessRespone MessageClass = 0b10
	ErrorResponse  MessageClass = 0b11
)

//     0                 1
//     2  3  4 5 6 7 8 9 0 1 2 3 4 5
//    +--+--+-+-+-+-+-+-+-+-+-+-+-+-+
//    |M |M |M|M|M|C|M|M|M|C|M|M|M|M|
//    |11|10|9|8|7|1|6|5|4|0|3|2|1|0|
//    +--+--+-+-+-+-+-+-+-+-+-+-+-+-+
//   "From RFC 8489 section-5 Fig 3"

func makeMessageType(method Method, mclass MessageClass) uint16 {
	mthd := uint16(method)
	m7_to_m11 := mthd & 0xf80 //   0b111110000000
	m4_to_m6 := mthd & 0x70   //   0b000001110000
	m0_to_m3 := mthd & 0xf    //   0b000000001111

	//shifting m4_to_m6 by 1 creates 1 bit space for c0 bit of class
	//shifting m7_to_m11 by 2 creates 1 bit space for c1 bit of class (2 shift because C0 also exists taking 1 space of the 14 bits )
	mthd = m0_to_m3 + (m4_to_m6 << 1) + (m7_to_m11 << 2)

	class := uint16(mclass)
	c0 := class & 0x1
	c1 := class & 0x2

	//place c0 at 5th pos and c1 at 8th pos
	class = (c0 << 4) + (c1 << 7)
	return mthd + class
}

func encodeMessageType(mtype uint16) [2]byte {
	return [2]byte{byte(mtype >> 8), byte(mtype << 8 >> 8)}
}

func isPayloadStun(p []byte) bool {

	if len(p) < 20 {
		return false
	}
	magic_cookie_bytes := p[4:8]

	return (magic_cookie_bytes[0] == byte(magic_cookie>>24) && magic_cookie_bytes[1] == byte(magic_cookie>>16) &&
		magic_cookie_bytes[2] == byte(magic_cookie>>8) && magic_cookie_bytes[3] == byte(magic_cookie))
}

// takes array of 2 bytes containing messageType inside stun header
func GetMessageType(messageType []byte) (Method, MessageClass) {

	mtype := uint16(messageType[0])<<8 | uint16(messageType[1])

	a := (mtype >> 2) & 0b111110000000
	b := (mtype >> 1) & 0b000001110000
	c := mtype & 0xF

	methodValue := a + b + c

	c0 := (mtype >> 4) & 0x1
	c1 := (mtype >> 7) & 0x2

	class := c0 + c1

	return Method(methodValue), MessageClass(class)
}

const stunSize int = 200
const udpSize int = 1400

func Process(conn *net.UDPConn) error {
	for {

		pb := make([]byte, udpSize+stunSize)
		_, addr, err := conn.ReadFromUDPAddrPort(pb)
		if !isPayloadStun(pb) {
			err = fmt.Errorf("Invalid Payload %x", pb)
			fmt.Println(err)
			return err
		}
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
		xor_map_attr_type := [2]byte{0, 0x20}
		xor_ipv4_attribute_length := [2]byte{0x0, 0x8}
		// xor_ipv6_attribute_length := [2]byte{0x0, 0x14}
		var ipv4_family byte = 0x01
		// var ipv6_family byte = 0x02
		var reserved byte = 0
		//			2 bytes (32bits)           2 bytes (32bits)
		// ___________________________________________________________
		//|			Attribute Type		|		Attribute Length	  |
		//|_____________________________|_____________________________|
		//|				|		     (Data)  	  					  |
		//|	  Reserved  |   Family  	|		Xor-Port		      |
		//|_____________|_______________|_____________________________|
		//|															  |
		//|						 Xor-Address						  |
		//| __________________________________________________________|
		//
		// Xor-Address is 4 bytes on ipv4 and 8 bytes on ipv6
		// Attribute type field is of 2 bytes 0x0020 represnts Xor-Mapped-Address Attribute

		// tp:=stun.NewType(stun.MethodBinding,stun.BindingSuccess.Class)
		// m2:=stun.NewWithOptions(stun.WithStrict())
		// (tp)
		// func xormapping(){

		// }
		//
		type_binding_response := encodeMessageType(makeMessageType(MethodBinding, SuccessRespone))
		reply_bytes := []byte{type_binding_response[0], type_binding_response[1], byte(0), 0x0c, mcb[0], mcb[1], mcb[2], mcb[3]}
		reply_bytes = append(reply_bytes, trcb...)
		reply_bytes = append(reply_bytes, xor_map_attr_type[:]...)
		reply_bytes = append(reply_bytes, xor_ipv4_attribute_length[:]...)
		reply_bytes = append(reply_bytes, reserved)
		reply_bytes = append(reply_bytes, ipv4_family)
		reply_bytes = append(reply_bytes, byte(xor_mapped_port<<8>>8), byte(xor_mapped_port>>8))
		reply_bytes = append(reply_bytes, byte(xor_mapped_address<<24>>24), byte(xor_mapped_address<<16>>24), byte(xor_mapped_address<<8>>24), byte(xor_mapped_address>>24))
		fmt.Println("stun?????", stun.IsMessage(reply_bytes))
		fmt.Printf("\n\nxor: %x %x\n", xor_mapped_address, xor_mapped_port)
		var send = stun.New()
		xmap := stun.XORMappedAddress{IP: net.ParseIP(addr.Addr().String())}
		xmap.AddTo(send)
		fmt.Printf("send %v", send.Attributes)
		_, err = conn.WriteToUDPAddrPort(reply_bytes, addr)
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
