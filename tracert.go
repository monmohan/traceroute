package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/monmohan/traceroute/icmp"
	"github.com/monmohan/traceroute/tcp"
)

func main() {
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	port := flag.Int("port", 80, "Port number when using TCP protocol")
	maxHops := flag.Int("maxHops", 64, "Maximum number of hops")
	proto := flag.String("proto", "icmp", "Protocol to use: 'tcp' or 'icmp'")
	iface := flag.String("iface", "any", `Interface to listen on. By default the program attempts to listens on all interfaces however it may not work on all platforms. 
	Provide the specific interface name if you face issues`)

	flag.Parse()
	if *maxHops < 1 {
		fmt.Println("Invalid number of hops, setting to default 64")
		*maxHops = 64
	}

	if *proto == "" || flag.NArg() < 1 {
		fmt.Println("Usage: tracert -proto [icmp|tcp] -verbose -port <port> -maxHops <maxHops> <Domin/IP address>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ipAddress := flag.Arg(0)
	// Resolve the IP address
	addr, err := net.ResolveIPAddr("ip4", ipAddress)
	if err != nil {
		fmt.Println("Failed to resolve IP address:", err)
		os.Exit(1)
	}
	fmt.Println("Resolved IP address:", addr)

	switch *proto {
	case "icmp":
		icmp.Trace(*verbose, *maxHops, addr)

	case "tcp":
		tcp.Trace(*iface, *verbose, *maxHops, addr, *port)

	default:
		fmt.Printf("Invalid Protocol specified: %s\n", *proto)
		os.Exit(1)
	}
}
