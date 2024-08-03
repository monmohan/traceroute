package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const timeout = time.Duration(20 * time.Second)

var verbose *bool

func debugPrint(v ...interface{}) {
	if *verbose {
		fmt.Println(v...)
	}
}

func main() {

	verbose = flag.Bool("verbose", false, "enable verbose output")
	maxHops := flag.Int("maxHops", 64, "number of hops")
	flag.Parse()
	if *verbose {
		fmt.Println("Verbose mode enabled")
	}

	if *maxHops < 1 {
		log.Println("Invalid number of hops, setting to default 64")
		*maxHops = 64
	}

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: tcptrace [ip_address]")
		return
	}
	// Get the IP address to trace

	ipAddr := args[0]

	// Resolve the IP address
	addr, err := net.ResolveIPAddr("ip4", ipAddr)
	if err != nil {
		fmt.Println("Failed to resolve IP address:", err)
		return
	}
	fmt.Println("Resolved IP address:", addr)

	//set up sync channels
	icmpChan := make(chan struct{})
	tcpChan := make(chan struct{})
	done := make(chan struct{})

	//go setUpICMPListener("en0", fmt.Sprintf("icmp and src host %s", addr.String()))
	go setUpICMPListener("en0", fmt.Sprintf("icmp or (tcp  and host %s)", addr.String()), icmpChan, tcpChan, done)
	go probe(addr, uint16(80), *maxHops, icmpChan, tcpChan, done)

	<-done
	fmt.Println("Done..")

}

func probe(addr *net.IPAddr, port uint16, maxHops int, icmpChan chan struct{}, tcpChan chan struct{}, done chan struct{}) error {
	//run probe with TTL 1 to 10
	for i := 1; i < maxHops; i++ {
		err := sendSyn(addr, port, i)
		if err != nil {
			debugPrint("Failed to probe:", err)

		}

		select {
		case icmpChan <- struct{}{}: // Signal ICMP request send
			debugPrint("TCP Probe: Signaled ICMP Channel")
		case <-time.After(timeout):
			debugPrint("TCP Send: Timeout while signaling ICMP Channel, continue probe")

		}
		debugPrint("\n--------------------------------------------------")
		select {
		case <-tcpChan: // Wait for ICMP Packet Read
			debugPrint("TCP Send: Received ICMP Channel Signal")
		case <-time.After(timeout):
			debugPrint("TCP Send: Timeout while waiting for ICMP Channel, continue to next probe")

		}

	}
	done <- struct{}{}
	return nil

}

func sendSyn(destIp *net.IPAddr, port uint16, ttl int) error {
	//ipConn, err := net.Dial("ip4:tcp", destIp.String())
	ipConn, err := net.DialIP("ip4:tcp", nil, &net.IPAddr{IP: destIp.IP})
	if err != nil {
		log.Fatal(err)
	}
	defer ipConn.Close()

	// Create IP layer
	ip := &layers.IPv4{
		SrcIP:    GetOutboundIP(),
		DstIP:    destIp.IP,
		Protocol: layers.IPProtocolTCP,
	}

	tcp := &layers.TCP{
		//Generate random port number each time
		SrcPort: layers.TCPPort(0xaa47 + uint16(ttl)),
		DstPort: layers.TCPPort(port),
		Seq:     rand.Uint32(),
		SYN:     true,
		Window:  65535,
		Urgent:  0,
		Options: []layers.TCPOption{},
	}
	tcp.SetNetworkLayerForChecksum(ip)

	// Serialize packet
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	/*
		Why not serialize IP packet along with TCP packet?
		When we serialized both IP and TCP layers, we were essentially creating an IP packet inside another IP packet.
		The outer IP packet (added by the OS) contained our entire serialized packet as its payload,
		leading to incorrect packet structure. Since the IP packet is created by the OS, we can't set its TTL direclty.
		Instead have to use raw sockets to set TTL.
	*/
	err = gopacket.SerializeLayers(buf, opts, tcp)
	if err != nil {
		return err
	}
	file, err := ipConn.File()
	if err != nil {
		return err
	}
	syscall.SetsockoptInt(int(file.Fd()), syscall.IPPROTO_IP, syscall.IP_TTL, ttl)

	_, err = ipConn.Write(buf.Bytes())

	if err != nil {
		return err
	}

	fmt.Println("Packet sent successfully with TTL ", ttl)

	return nil

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

func setUpICMPListener(dev string, filter string, icmpChan chan struct{}, tcpChan chan struct{}, done chan struct{}) {
	handle, err := pcap.OpenLive(dev, 1600, false, pcap.BlockForever)
	// print what is captured
	debugPrint("Capturing packets on interface", dev)

	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()
	// Set BPF filter
	err = handle.SetBPFFilter(filter)
	debugPrint("Filter set to", filter)
	if err != nil {
		log.Fatal(err)
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for {
		select {
		case <-icmpChan: // Wait for TCP request send
			debugPrint("ICMP Listener: Received TCP Channel Signal")
		case <-time.After(timeout):
			debugPrint("ICMP Listener: Timeout while waiting for TCP Send")
			done <- struct{}{}
			return
		}
		debugPrint("ICMP Listener: Trying to get next Packet")
		packet, err := packetSource.NextPacket()
		debugPrint("ICMP Listener: Received Packet")
		if err != nil {
			debugPrint("Failed to get next packet:", err)
		}
		//toggle on layer type
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			if isTCPAck(packet) {
				{
					debugPrint("ICMP Listener: Got TCP ACK Packet, we are done")
					done <- struct{}{}
					return
				}
			}
			debugPrint("ICMP Listener: Continue to wait for ICMP Packet")
			packet, err = packetSource.NextPacket()
			if err != nil {
				debugPrint("Failed to get next packet:", err)
			}
			debugPrint("ICMP Listener: Received Packet")

		}

		if src := getICMPInfo(packet); src != "" {
			fmt.Println(src)
		}

		select {
		case tcpChan <- struct{}{}: // Signal TCP request send
			debugPrint("ICMP Listener: Signaled TCP Channel")
		case <-time.After(timeout):
			debugPrint("ICMP Listener: Timeout while signaling TCP Channel")
			done <- struct{}{}
			return
		}

	}

}

func isTCPAck(packet gopacket.Packet) bool {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		debugPrint(packet)
		//print syn ISN
		debugPrint("SYN ISN: ", tcp.Seq)
		//print ack ISN
		debugPrint("ACK ISN: ", tcp.Ack)

		return tcp.ACK && tcp.SYN
	}
	return false
}
func getICMPInfo(packet gopacket.Packet) string {
	// Let's see if the packet is an ICMP packet
	icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
	if icmpLayer != nil {
		debugPrint("ICMP packet detected")

		icmp, _ := icmpLayer.(*layers.ICMPv4)

		src := packet.NetworkLayer().NetworkFlow().Src()

		debugPrint(fmt.Sprintf("From %v to %v\n",
			src,
			packet.NetworkLayer().NetworkFlow().Dst()))

		debugPrint("ICMP Type: ", icmp.TypeCode.Type())
		debugPrint("ICMP Code: ", icmp.TypeCode.Code())

		// Print more details based on ICMP type
		switch icmp.TypeCode.Type() {
		case layers.ICMPv4TypeEchoRequest, layers.ICMPv4TypeEchoReply:
			debugPrint("ICMP ID: ", icmp.Id)
			debugPrint("ICMP Sequence: ", icmp.Seq)
		case layers.ICMPv4TypeDestinationUnreachable:
			debugPrint("Destination Unreachable")
		case layers.ICMPv4TypeTimeExceeded:
			debugPrint("Time Exceeded")
		}

		debugPrint("--- End of ICMP Packet ---")
		return fmt.Sprintf("%v", src)
	}
	return ""

}
