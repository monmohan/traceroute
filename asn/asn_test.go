package asn

import (
	"encoding/csv"
	"fmt"
	"os"
	"testing"
)

func TestNewASN(t *testing.T) {
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
	asnObj := NewASNTree(reader)

	// check if the returned object is not nil
	if asnObj == nil {
		t.Fatal("ASN object is nil")
	}

}

func TestPrintASN(t *testing.T) {
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

}

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
	asnObj := NewASNTree(reader)

	// check if the returned object is not nil
	if asnObj == nil {
		t.Fatal("ASN object is nil")
	}

	// call FindASN with a valid IP address
	// check if the returned ASNData is not nil
	ip := "1.5.140.0"
	asnData, err := asnObj.FindASN(ip)
	if err != nil {
		t.Fatalf("failed to find ASN: %v", err)
	}
	fmt.Println(asnData)

}
