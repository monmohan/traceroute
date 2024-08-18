package asn

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
)

type ASNQuery interface {
	FindASN(ip string) (ASNData, error)
}
type ASNTree struct {
	value    byte
	children []ASNTree
	isRoot   bool
	isLeaf   bool
	data     ASNData
}
type ASNData struct {
	IPStart, IPEnd net.IP
	ASNNumber      string
	CountryCode    string
	ASName         string
}

/**
Expected format of the CSV file
IPStart\tIPEnd\tASNNumber\tCountryCode\tASName
*/

func NewASNTree(input *csv.Reader) *ASNTree {
	ASN := &ASNTree{value: 0, children: make([]ASNTree, 0), isRoot: true}

	FindOrAppend := func(tree *ASNTree, value byte) *ASNTree {
		for i, v := range tree.children {
			if v.value == value {
				return &tree.children[i]
			}
		}
		// Value not found, append it
		node := ASNTree{value: value, children: make([]ASNTree, 0)}
		tree.children = append(tree.children, node)
		return &tree.children[len(tree.children)-1]
	}

	head := 200
	counter := 0
	readAll := true
	for (counter < head) || readAll {
		record, err := input.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Error reading record: ", err)
		}
		data := ASNData{IPStart: net.ParseIP(record[0]),
			IPEnd:       net.ParseIP(record[1]).To4(),
			ASNNumber:   record[2],
			CountryCode: record[3],
			ASName:      record[4]}

		// Parse the data
		fromIP := data.IPStart.To4()
		child := FindOrAppend(ASN, fromIP[0])
		for i := 1; i < len(fromIP); i++ {
			child = FindOrAppend(child, fromIP[i])
		}
		child.isLeaf = true
		child.data = data

		//check if the IP in map
		counter++

	}

	//print the tree recursively
	//PrintASNTree(*ASN, "", true)
	return ASN

}
func PrintASNTree(node ASNTree, prefix string, isLast bool) {
	// Print the current node
	if node.isRoot {
		fmt.Println("ASNTree (Root)")
	} else {
		if isLast {
			fmt.Printf("%s└── ", prefix)
			prefix += "    "
		} else {
			fmt.Printf("%s├── ", prefix)
			prefix += "│   "
		}
		if node.isLeaf {
			fmt.Printf("Value: %d, data %s\n", node.value, node.data)
		} else {
			fmt.Printf("Value: %d\n", node.value)
		}

	}

	// Print children
	for i, child := range node.children {
		isLastChild := i == len(node.children)-1
		PrintASNTree(child, prefix, isLastChild)
	}
}

func (asn *ASNTree) FindASN(ip string) (ASNData, error) {
	//parse the IP
	ipAddr := net.ParseIP(ip).To4()
	if ipAddr == nil {
		return ASNData{}, fmt.Errorf("invalid IP Address: %s", ip)
	}
	return asn.findASN(ipAddr, 0)

}

func (asn *ASNTree) findASN(ip net.IP, start int) (ASNData, error) {
	if asn.isLeaf {
		if isInRange(ip, asn.data.IPStart, asn.data.IPEnd) {
			return asn.data, nil
		}
		return ASNData{}, fmt.Errorf("IP not found in the tree")
	}
	child := findChild(asn.children, ip[start])
	if child == -1 {
		return ASNData{}, fmt.Errorf("IP not found in the tree")
	}

	return asn.children[child].findASN(ip, start+1)
}

func isInRange(ip net.IP, start net.IP, end net.IP) bool {
	if ip[0] < start[0] || ip[0] > end[0] {
		return false
	}
	for i := 1; i < len(ip); i++ {
		if ip[i] < start[i] || ip[i] > end[i] {
			return false
		}
	}
	return true
}

func findChild(list []ASNTree, sval byte) int {

	for i, v := range list {
		if i == 0 && v.value > sval {
			return -1
		}
		if i == len(list)-1 && v.value < sval {
			return i
		}
		if v.value == sval {
			return i
		}
		if v.value > sval {
			if i+1 < len(list) && list[i+1].value > sval {
				return i
			}

		}
	}
	return -1
}
