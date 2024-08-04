# traceroute
Go implementation and exploration of traceroute using TCP and ICMP

# Running an ICMP Trace
```go

$ sudo go run tracert.go -proto icmp accounts.google.com
Invalid number of hops, setting to default 64
Resolved IP address: 142.251.175.84
Sent ICMP probe to 142.251.175.84 with TTL 1 Time exceeded from peer  192.168.18.1
Sent ICMP probe to 142.251.175.84 with TTL 2 Time exceeded from peer  116.88.128.1
Sent ICMP probe to 142.251.175.84 with TTL 3 Time exceeded from peer  183.90.44.189
Sent ICMP probe to 142.251.175.84 with TTL 4 Time exceeded from peer  203.118.6.233
Sent ICMP probe to 142.251.175.84 with TTL 5 Time exceeded from peer  203.118.6.149
Sent ICMP probe to 142.251.175.84 with TTL 6 Time exceeded from peer  203.118.4.130
Sent ICMP probe to 142.251.175.84 with TTL 7 Time exceeded from peer  142.250.166.50
Sent ICMP probe to 142.251.175.84 with TTL 8 Time exceeded from peer  142.250.238.117
Sent ICMP probe to 142.251.175.84 with TTL 9 Time exceeded from peer  142.250.60.240
Sent ICMP probe to 142.251.175.84 with TTL 10 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 11 Time exceeded from peer  142.251.231.198
Sent ICMP probe to 142.251.175.84 with TTL 12 Time exceeded from peer  142.251.247.195
Sent ICMP probe to 142.251.175.84 with TTL 13 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 14 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 15 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 16 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 17 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 18 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 19 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 20 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 21 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 22 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 23 Failed to receive ICMP reply
Sent ICMP probe to 142.251.175.84 with TTL 24 Echo reply from peer  142.251.175.84
Done..

```
As you can see the packet took 24 hops to reach its destination accounts.google.com and we are able to see the IPs of different routers (e.g. 203.118.6.233) when they send time exceeded ICMP message. Many routers didn't respond and once we get final echo reply from destination, the trace ends

# Running a TCP trace
```go
$ sudo go run tracert.go -proto tcp accounts.google.com
Invalid number of hops, setting to default 64
Resolved IP address: 74.125.68.84
Packet sent with TTL : 1  ICMP Packet Received from :  192.168.18.1
Packet sent with TTL : 2  ICMP Packet Received from :  116.88.128.1
Packet sent with TTL : 3  ICMP Packet Received from :  183.90.44.193
Packet sent with TTL : 4  ICMP Packet Received from :  203.118.6.237
Packet sent with TTL : 5  ICMP Packet Received from :  203.118.6.149
Packet sent with TTL : 6  ICMP Packet Received from :  203.118.6.149
Packet sent with TTL : 7  ICMP Packet Received from :  203.118.4.130
Packet sent with TTL : 8  ICMP Packet Received from :  142.250.166.50
Packet sent with TTL : 9  ICMP Packet Received from :  142.250.238.115
Packet sent with TTL : 10  * * * Timeout while waiting for ICMP Packet * * * 
Packet sent with TTL : 11  ICMP Packet Received from :  209.85.255.43
Packet sent with TTL : 12  ICMP Packet Received from :  216.239.35.171
Packet sent with TTL : 13  ICMP Packet Received from :  108.170.234.59
Packet sent with TTL : 14  * * * Timeout while waiting for ICMP Packet * * * 
Packet sent with TTL : 15  * * * Timeout while waiting for ICMP Packet * * * 
Packet sent with TTL : 16  * * * Timeout while waiting for ICMP Packet * * * 
Packet sent with TTL : 17  * * * Timeout while waiting for ICMP Packet * * * 
Packet sent with TTL : 18  * * * Timeout while waiting for ICMP Packet * * * 
Packet sent with TTL : 19  * * * Timeout while waiting for ICMP Packet * * * 
Packet sent with TTL : 20  * * * Timeout while waiting for ICMP Packet * * * 
Packet sent with TTL : 21  * * * Timeout while waiting for ICMP Packet * * * 
Packet sent with TTL : 22  * * * Timeout while waiting for ICMP Packet * * * 
Packet sent with TTL : 23 Got TCP ACK Packet from :  74.125.68.84
Done..

```
Here the results are pretty similar except that we are sending a TCP SYN and waiting for either an ICMP Time Exceeded or an ACK from the destination. Again, the packet took 24 hops to reach its destination accounts.google.com and we are able to see the IPs of different routers (e.g. 209.85.255.43) when they send time exceeded ICMP message. Many routers didn't respond and once we get TCP ACK from destination, the trace ends
