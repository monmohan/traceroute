package main

import (
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

func main() {
	// Get the IP address to ping
	if len(os.Args) != 2 {
		fmt.Println("Usage: ping [ip_address]")
		return
	}
	ipAddr := os.Args[1]

	// Resolve the IP address
	addr, err := net.ResolveIPAddr("ip4", ipAddr)
	if err != nil {
		fmt.Println("Failed to resolve IP address:", err)
		return
	}
	// Open a raw socket for sending/receiving ICMP messages
	conn, err := setUpICMPListener()
	if err != nil {
		fmt.Println("Failed to open ICMP socket:", err)
		return
	}

	defer conn.Close()

	ttl := 1

	//run in loop until destination is reached or max TTL is reached
	for ttl <= 30 {
		peer, err := runICMPProbe(conn, ttl, addr)
		if err != nil {
			fmt.Println("Failed to probe:", err)

		}
		if peer != nil && peer.String() == addr.String() {
			fmt.Println("Reached destination, hops needed to reach destination:", ttl)
			break
		}
		ttl++
	}

}

func runICMPProbe(conn *icmp.PacketConn, ttl int, addr *net.IPAddr) (net.Addr, error) {
	conn.IPv4PacketConn().SetTTL(ttl)

	icmpMsg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("Go Ping"),
		},
	}

	msg, err := icmpMsg.Marshal(nil)
	if err != nil {
		return nil, err
	}

	if _, err = conn.WriteTo(msg, addr); err != nil {
		return nil, fmt.Errorf("failed to send ICMP message: %v", err)

	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	return readICMPResponse(conn)

}

func setUpICMPListener() (*icmp.PacketConn, error) {
	// Open a raw socket for ICMP messages
	conn, err := icmp.ListenPacket("ip4:icmp", GetOutboundIP().String())
	if err != nil {
		fmt.Println("Failed to open ICMP socket:", err)
		return nil, err
	}
	return conn, nil
}

func readICMPResponse(conn *icmp.PacketConn) (net.Addr, error) {
	reply := make([]byte, 1500)
	n, peer, err := conn.ReadFrom(reply)
	if err != nil {
		return nil, fmt.Errorf("failed to receive ICMP reply: %v", err)

	}

	packet := gopacket.NewPacket(reply[:n], layers.LayerTypeICMPv4, gopacket.Default)
	icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
	if icmpLayer == nil {
		return nil, fmt.Errorf("failed to parse ICMP reply")

	}
	icmpPacket, _ := icmpLayer.(*layers.ICMPv4)
	fmt.Printf("Reply from %v: ", peer)

	fmt.Printf("Type: %v, Code: %v, ID: %v, Seq: %v, ", icmpPacket.TypeCode.Type(), icmpPacket.TypeCode.Code(), icmpPacket.Id, icmpPacket.Seq)

	switch icmpPacket.TypeCode.Type() {
	case layers.ICMPv4TypeEchoReply:
		fmt.Println("Echo reply")
	case layers.ICMPv4TypeDestinationUnreachable:
		fmt.Println("Destination unreachable")
	case layers.ICMPv4TypeTimeExceeded:
		fmt.Println("Time exceeded")
	default:
		fmt.Println("Unknown")
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
