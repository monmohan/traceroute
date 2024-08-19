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
type ASNData struct {
	IPStart, IPEnd net.IP
	ASNNumber      string
	CountryCode    string
	ASName         string
}

type Reader interface {
	ReadAll(input *csv.Reader) ASNQuery
}

type RangeReader struct {
	FromIPs []int
	ToIPs   []int
	ASNData []ASNData
}

/*
*
Expected format of the CSV file
IPStart\tIPEnd\tASNNumber\tCountryCode\tASName
*/
func (rangeReader *RangeReader) ReadAll(input *csv.Reader) {
	for {
		record, err := input.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Error reading record: ", err)
		}
		rangeReader.handleRecord(record)

	}

}
func NewRangeReader() *RangeReader {
	return &RangeReader{FromIPs: make([]int, 0), ToIPs: make([]int, 0), ASNData: make([]ASNData, 0)}
}

func (rr *RangeReader) handleRecord(record []string) {
	data := ASNData{IPStart: net.ParseIP(record[0]),
		IPEnd:       net.ParseIP(record[1]).To4(),
		ASNNumber:   record[2],
		CountryCode: record[3],
		ASName:      record[4]}

	rr.FromIPs = append(rr.FromIPs, ToInt(data.IPStart))
	rr.ToIPs = append(rr.ToIPs, ToInt(data.IPEnd))
	rr.ASNData = append(rr.ASNData, data)

}

func ToInt(ip net.IP) int {
	ip = ip.To4()
	var ret int = 0
	for i := 0; i < len(ip); i++ {
		t := int(ip[i])
		t <<= 8 * (3 - i)
		ret = ret | t
	}
	return ret
}

func (rr *RangeReader) FindASN(ip string) (ASNData, error) {
	//parse the IP
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return ASNData{}, fmt.Errorf("invalid IP address: %s", ip)
	}
	in := ToInt(ipAddr)
	len := len(rr.FromIPs)
	for i := 0; i < len; i++ {
		if in >= rr.FromIPs[i] && in <= rr.ToIPs[i] {
			return rr.ASNData[i], nil
		}
	}
	return ASNData{}, fmt.Errorf("no ASN found: %s", ip)
}

////////////// ----This reader implementation was experiment. RangeReader is the correct implementation//////////////
////////////// ----Keeping it commented------------------------------//////////////

/*type ASNTree struct {
	value    byte
	children []ASNTree
	isRoot   bool
	isLeaf   bool
	data     ASNData
}*/

// The TreeReader struct represents a reader that uses a tree structure to store and query ASN data.
// type TreeReader struct {
// 	ASNTree *ASNTree
// }

// NewTreeReader creates a new instance of TreeReader with an initialized ASNTree.
// func NewTreeReader() *TreeReader {
// 	ASN := &ASNTree{value: 0, children: make([]ASNTree, 0), isRoot: true}
// 	return &TreeReader{ASNTree: ASN}
// }

// ReadAll reads all records from the input CSV reader and populates the ASNTree with the data.
// func (treeReader *TreeReader) ReadAll(input *csv.Reader) {
// 	for {
// 		record, err := input.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			log.Fatal("Error reading record: ", err)
// 		}
// 		treeReader.handleRecord(record)
// 	}
// }

// FindOrAppend finds a child node with the specified value in the given tree, or appends a new child node if not found.
// func FindOrAppend(tree *ASNTree, value byte) *ASNTree {
// 	for i, v := range tree.children {
// 		if v.value == value {
// 			return &tree.children[i]
// 		}
// 	}
// 	// Value not found, append it
// 	node := ASNTree{value: value, children: make([]ASNTree, 0)}
// 	tree.children = append(tree.children, node)
// 	return &tree.children[len(tree.children)-1]
// }

// handleRecord processes a single record from the CSV file and adds it to the ASNTree.
// func (treeReader *TreeReader) handleRecord(record []string) {
// 	tree := treeReader.ASNTree
// 	data := ASNData{
// 		IPStart:     net.ParseIP(record[0]),
// 		IPEnd:       net.ParseIP(record[1]).To4(),
// 		ASNNumber:   record[2],
// 		CountryCode: record[3],
// 		ASName:      record[4],
// 	}

// 	// Parse the data
// 	fromIP := data.IPStart.To4()
// 	child := FindOrAppend(tree, fromIP[0])
// 	for i := 1; i < len(fromIP); i++ {
// 		child = FindOrAppend(child, fromIP[i])
// 	}
// 	child.isLeaf = true
// 	child.data = data
// }

// PrintASNTree prints the ASNTree structure in a tree-like format.
// It recursively traverses the tree and prints each node along with its value and data (if available).
// func PrintASNTree(node ASNTree, prefix string, isLast bool) {
// 	// Print the current node
// 	if node.isRoot {
// 		fmt.Println("ASNTree (Root)")
// 	} else {
// 		if isLast {
// 			fmt.Printf("%s└── ", prefix)
// 			prefix += "    "
// 		} else {
// 			fmt.Printf("%s├── ", prefix)
// 			prefix += "│   "
// 		}
// 		if node.isLeaf {
// 			fmt.Printf("Value: %d, data %s\n", node.value, node.data)
// 		} else {
// 			fmt.Printf("Value: %d\n", node.value)
// 		}
// 	}

// 	// Print children
// 	for i, child := range node.children {
// 		isLastChild := i == len(node.children)-1
// 		PrintASNTree(child, prefix, isLastChild)
// 	}
// }

// FindASN finds the ASNData associated with the given IP address in the ASNTree.
// func (treeReader *TreeReader) FindASN(ip string) (ASNData, error) {
// 	// Parse the IP
// 	ipAddr := net.ParseIP(ip).To4()
// 	if ipAddr == nil {
// 		return ASNData{}, fmt.Errorf("invalid IP Address: %s", ip)
// 	}
// 	return treeReader.ASNTree.findASN(ipAddr, 0)
// }

// findASN recursively searches the ASNTree for the ASNData associated with the given IP address.
// func (asn *ASNTree) findASN(ip net.IP, start int) (ASNData, error) {
// 	if asn.isLeaf {
// 		if isInRange(ip, asn.data.IPStart, asn.data.IPEnd) {
// 			return asn.data, nil
// 		}
// 		return ASNData{}, fmt.Errorf("IP not found in the tree")
// 	}
// 	child := findChild(asn.children, ip[start])
// 	if child == -1 {
// 		return ASNData{}, fmt.Errorf("IP not found in the tree")
// 	}
// 	return asn.children[child].findASN(ip, start+1)
// }

// isInRange checks if the given IP address is within the range specified by start and end IP addresses.
// func isInRange(ip net.IP, start net.IP, end net.IP) bool {
// 	if ip[0] < start[0] || ip[0] > end[0] {
// 		return false
// 	}
// 	for i := 1; i < len(ip); i++ {
// 		if ip[i] < start[i] || ip[i] > end[i] {
// 			return false
// 		}
// 	}
// 	return true
// }

// findChild finds the index of the child node with the specified value in the given list of ASNTree nodes.
// func findChild(list []ASNTree, sval byte) int {
// 	for i, v := range list {
// 		if i == 0 && v.value > sval {
// 			return -1
// 		}
// 		if i == len(list)-1 && v.value < sval {
// 			return i
// 		}
// 		if v.value == sval {
// 			return i
// 		}
// 		if v.value > sval {
// 			if i+1 < len(list) && list[i+1].value > sval {
// 				return i
// 			}
// 		}
// 	}
// 	return -1
// }
