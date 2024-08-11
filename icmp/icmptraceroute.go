package icmp

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var dbg bool

func debugPrint(v ...interface{}) {
	if dbg {
		fmt.Println(v...)
	}
}

func Trace(verbose bool, maxHops int, ipAddr *net.IPAddr) {
	dbg = verbose
	flag.Parse()

	laddr := GetOutboundIP()

	for ttl := 1; ttl <= maxHops; ttl++ {
		debugPrint("-------------------Start Probe with TTL ", ttl, "-------------------")
		retAddr, err := runICMPProbe(ipAddr, laddr, ttl)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if retAddr.String() == ipAddr.IP.String() {
			break
		}
		debugPrint("-------------------End Probe with TTL ", ttl, "-------------------\n")
	}

	fmt.Println("Done..")

}

func runICMPProbe(addr *net.IPAddr, laddr net.IP, ttl int) (net.Addr, error) {

	conn, err := icmp.ListenPacket("ip4:icmp", laddr.String())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.IPv4PacketConn().SetTTL(ttl)
	echoRequest := &icmp.Echo{
		ID:   os.Getpid() & 0xffff,
		Seq:  ttl, //incremented in each iteration
		Data: []byte("PING.."),
	}

	icmpMsg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		/*
		 RFC 792 (Internet Control Message Protocol),
		 the ID and Sequence fields in an ICMP Echo Request/Reply are 16-bit fields, not 32-bit.
		 Over the wire they will sent out as 16-bit fields.
		 Later we will retrieve them as such
		*/
		Body: echoRequest,
	}

	msg, err := icmpMsg.Marshal(nil)
	if err != nil {
		return nil, err
	}

	if _, err = conn.WriteTo(msg, addr); err != nil {
		return nil, fmt.Errorf("failed to send ICMP message: %v", err)

	}
	fmt.Println("Sent ICMP Echo Request to ", addr, " with TTL/Seq ", ttl)

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	return readICMPResponse(conn, echoRequest)

}

func readICMPResponse(conn *icmp.PacketConn, echoRequest *icmp.Echo) (net.Addr, error) {
	reply := make([]byte, 1500)
	n, peer, err := conn.ReadFrom(reply)
	if err != nil {
		debugPrint(`failed to receive ICMP reply:`, err)
		return nil, fmt.Errorf("failed to receive ICMP reply")

	}

	packet := gopacket.NewPacket(reply[:n], layers.LayerTypeICMPv4, gopacket.Default)
	icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
	if icmpLayer == nil {
		return nil, fmt.Errorf("failed to parse ICMP reply")

	}
	icmpPacket, _ := icmpLayer.(*layers.ICMPv4)
	debugPrint("Reply from : ", peer)
	/**

		 RFC 792
		Time Exceeded Message

	    0                   1                   2                   3
	    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |     Type      |     Code      |          Checksum             |
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |                             unused                            |
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |      Internet Header + 64 bits of Original Data Datagram      |
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

		**/

	switch icmpPacket.TypeCode.Type() {
	case layers.ICMPv4TypeEchoReply:
		fmt.Println("Echo reply from peer ", peer)
		// For Echo Reply, ID and Seq are directly available in the ICMP header
		if icmpPacket.Id == uint16(echoRequest.ID) && icmpPacket.Seq == uint16(echoRequest.Seq) {
			debugPrint("Found original message in Echo Reply, ID and Sequence match\n")
		} else {
			fmt.Println("IGNORE: Echo Reply does not match original message")
		}

	case layers.ICMPv4TypeDestinationUnreachable:
		debugPrint("Destination unreachable")
	case layers.ICMPv4TypeTimeExceeded:
		fmt.Println("Time exceeded from peer ", peer)

		if len(icmpPacket.Payload) >= 28 { // 20 bytes IP header + 8 bytes original ICMP header
			// Get the IP header length
			ipHeaderLength := int(icmpPacket.Payload[0]&0x0f) * 4
			debugPrint("IP Header Length: ", ipHeaderLength)

			originalICMP := icmpPacket.Payload[ipHeaderLength:] // Skip the IP header
			// 8 is the type for Echo Request
			if originalICMP[0] == 8 {
				/*
					Network byte order is, by convention, big-endian. RFC 1700
				*/
				code := originalICMP[1] //should be 0
				checksum := binary.BigEndian.Uint16(originalICMP[2:4])
				id := binary.BigEndian.Uint16(originalICMP[4:6])
				seq := binary.BigEndian.Uint16(originalICMP[6:8])
				debugPrint(fmt.Sprintf("Data read from ICMP Response => Code: %d, Checksum: %d, ID: %d, Sequence: %d", code, checksum, id, seq))
				if id == uint16(echoRequest.ID) && seq == uint16(echoRequest.Seq) {
					debugPrint(fmt.Sprintf("Found original message in Time Exceeded payload, ID %d and Sequence %d match\n ", id, seq))
				} else {
					fmt.Println("IGNORE: Time Exceeded payload does not match original message")
				}

			} else {
				fmt.Println("Original message was not an Echo Request")
			}
		} else {
			fmt.Println("Time Exceeded payload too short to extract original message")
		}

	default:
		debugPrint("Unknown")
	}

	return peer, nil

}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
