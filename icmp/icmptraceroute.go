package icmp

import (
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
		retAddr, err := runICMPProbe(ipAddr, laddr, ttl)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if retAddr.String() == ipAddr.IP.String() {
			break
		}
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

	icmpMsg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("PING.."),
		},
	}

	msg, err := icmpMsg.Marshal(nil)
	if err != nil {
		return nil, err
	}

	if _, err = conn.WriteTo(msg, addr); err != nil {
		return nil, fmt.Errorf("failed to send ICMP message: %v", err)

	}
	fmt.Print("Sent ICMP probe to ", addr, " with TTL ", ttl, " ")

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	return readICMPResponse(conn)

}

func readICMPResponse(conn *icmp.PacketConn) (net.Addr, error) {
	reply := make([]byte, 1500)
	n, peer, err := conn.ReadFrom(reply)
	if err != nil {
		debugPrint(`failed to receive ICMP reply:`, err)
		return nil, fmt.Errorf("Failed to receive ICMP reply")

	}

	packet := gopacket.NewPacket(reply[:n], layers.LayerTypeICMPv4, gopacket.Default)
	icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
	if icmpLayer == nil {
		return nil, fmt.Errorf("failed to parse ICMP reply")

	}
	icmpPacket, _ := icmpLayer.(*layers.ICMPv4)
	debugPrint("Reply from : ", peer)

	debugPrint(fmt.Sprintf("Type: %v, Code: %v, ID: %v, Seq: %v, ", icmpPacket.TypeCode.Type(), icmpPacket.TypeCode.Code(), icmpPacket.Id, icmpPacket.Seq))

	switch icmpPacket.TypeCode.Type() {
	case layers.ICMPv4TypeEchoReply:
		fmt.Println("Echo reply from peer ", peer)
	case layers.ICMPv4TypeDestinationUnreachable:
		debugPrint("Destination unreachable")
	case layers.ICMPv4TypeTimeExceeded:
		fmt.Println("Time exceeded from peer ", peer)
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
