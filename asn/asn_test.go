package asn

import (
	"encoding/csv"
	"net"
	"os"
	"testing"
)

func TestNewTreeReader(t *testing.T) {
	// write test case for NewASN
	// read the testdata/asn.csv file
	file, err := os.Open("/Users/singhmo/CSPrimer/Networking/IP_ICMP/ip2asn-v4.tsv")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// create a csv.Reader
	reader := csv.NewReader(file)
	reader.Comma = '\t'

	// call NewASN with the csv.Reader
	asnObj := NewRangeReader()
	asnObj.ReadAll(reader)

	// check if the returned object is not nil
	if asnObj == nil {
		t.Fatal("ASN object is nil")
	}

}

/*func TestPrintASN(t *testing.T) {
	tree := ASNTree{
		value:  1,
		isRoot: true,
		children: []ASNTree{
			{value: 2, children: []ASNTree{}},
			{value: 3, children: []ASNTree{}},
			{value: 11, children: []ASNTree{}},
		},
	}
	PrintASNTree(tree, "", true)

}*/

func TestFindASN(t *testing.T) {
	// write test case for FindASN
	// read the testdata/asn.csv file
	file, err := os.Open("/Users/singhmo/CSPrimer/Networking/IP_ICMP/ip2asn-v4.tsv")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// create a csv.Reader
	reader := csv.NewReader(file)
	reader.Comma = '\t'

	// call NewASN with the csv.Reader
	asnObj := NewRangeReader()
	asnObj.ReadAll(reader)

	// check if the returned object is not nil
	if asnObj == nil {
		t.Fatal("ASN object is nil")
	}

	// call FindASN with a valid IP address
	// check if the returned ASNData is not nil
	//add test for multiple IPs

	tests := []struct {
		ip   string
		want ASNData
	}{
		{

			ip: "1.5.140.0",
			want: ASNData{
				IPStart:     net.ParseIP("1.5.0.0"),
				IPEnd:       net.ParseIP("1.5.255.255"),
				ASNNumber:   "4725",
				CountryCode: "JP",
				ASName:      "ODN SoftBank Corp.",
			},
		},
		{
			ip: "203.118.7.77",
			want: ASNData{
				IPStart:     net.ParseIP("203.117.254.0"),
				IPEnd:       net.ParseIP("203.118.10.255"),
				ASNNumber:   "4657",
				CountryCode: "SG",
				ASName:      "STARHUB-INTERNET StarHub Ltd",
			},
		},
		{

			ip: "207.45.219.137",
			want: ASNData{
				IPStart:     net.ParseIP("207.45.192.0"),
				IPEnd:       net.ParseIP("207.45.223.255"),
				ASNNumber:   "6453",
				CountryCode: "US",
				ASName:      "AS6453",
			},
		},
		{
			ip: "183.90.44.189",
			want: ASNData{
				IPStart:     net.ParseIP("183.90.40.0"),
				IPEnd:       net.ParseIP("183.90.127.255"),
				ASNNumber:   "55430",
				CountryCode: "SG",
				ASName:      "STARHUB-NGNBN Starhub Ltd",
			},
		},
		{
			ip: "99.83.65.110",
			want: ASNData{
				IPStart:     net.ParseIP("99.83.64.0"),
				IPEnd:       net.ParseIP("99.83.71.255"),
				ASNNumber:   "0",
				CountryCode: "None",
				ASName:      "Not routed",
			},
		},
		{
			ip: "108.170.234.57",
			want: ASNData{
				IPStart:     net.ParseIP("108.170.224.0"),
				IPEnd:       net.ParseIP("108.170.255.255"),
				ASNNumber:   "15169",
				CountryCode: "US",
				ASName:      "GOOGLE",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.ip, func(t *testing.T) {
			got, err := asnObj.FindASN(test.ip)
			if err != nil {
				t.Fatalf("failed to find ASN: %v", err)
			}
			if got.ASNNumber != test.want.ASNNumber && got.ASName != test.want.ASName && got.CountryCode != test.want.CountryCode {
				t.Errorf("got: %v, want: %v", got, test.want)
			}
		})
	}

	/*--Just printing the ASNData for the IPs--*/
	/*ips := []string{"1.5.140.0", "203.118.7.77", "207.45.219.137",
		"216.6.87.227", "183.90.44.189",
		"142.251.230.137", "99.83.65.110", "108.170.234.57", "180.87.106.0"}

	for _, ip := range ips {
		fmt.Println("Finding ASN for IP:", ip)

		asnData, err := asnObj.FindASN(ip)
		if err != nil {
			t.Fatalf("failed to find ASN: %v", err)
		}
		fmt.Println(asnData)
	}*/

}
