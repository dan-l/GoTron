package main

import (
	"fmt"
	//"log"
	"net"
	//"net/rpc"
	"encoding/json"
	"os"
	"sync"
)

// main context
type KeyValService struct {
	NodeLock sync.RWMutex
	NodeList map[string]net.Conn
	// lame solution, k = node, v = keys
	KeyLocations map[string]map[string]struct{} // note: a map[string]struct{} acts list a Set<String> from Java

	MessageId int32 // atomically incremented message id

	UnavailableKeys map[string]struct{}
}

type Message struct {
	Id      int32
	NodeId  string
	Message string
}

// check if a fatal error has ocurred
func FatalError(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(-10)
	}
}

// Level for printing
// 0 - only errors
// 1 - general connection info, key info
// 2 - message aggreagtion
// 3 - Messages being sent
// 4 - Everything
const DebugLevel int = 1

func DebugPrint(level int, str string) {
	if level <= DebugLevel {
		fmt.Println(str)
	}
}

type HelloMessage struct {
	Id              string
	Keys            []string
	UnavailableKeys []string
}

// this is called when a node joins, it handles adding the node to lists and adding the keys it knows about
func AddNode(this *KeyValService, hello *HelloMessage, conn net.Conn) {

}

func SplitMessages(raw []byte) [][]byte {
	ret := make([][]byte, 0)

	base := -1
	for index, val := range raw {
		if val == '{' {
			base = index
		} else if val == '}' {
			ret = append(ret, raw[base:index+1])
		}
	}

	return ret
}

func HandleMessage(this *KeyValService, message Message) {
}

// called when a new node connects, and processes responses from it
func HandleConnect(this *KeyValService, conn net.Conn) {
	buffer := make([]byte, 1024)

	// reading the hello message
	n, e := conn.Read(buffer)
	if e != nil {
		fmt.Println("HandleConnect", e)
		return
	}
	// process the hello message
	var hello HelloMessage
	e = json.Unmarshal(buffer[0:n], &hello)
	if e != nil {
		fmt.Println("Unmarshal hello:", e)
		return
	}

	DebugPrint(3, "KV->FE: "+string(buffer))

	// store data locally
	AddNode(this, &hello, conn)

	DebugPrint(1, "SUCCESS")
}

// the listening loop
func ListenConnections(this *KeyValService, addr string) {
	conn, e := net.Listen("tcp", addr)
	defer conn.Close()
	FatalError(e)

	for {
		newConn, e := conn.Accept()
		if e != nil {
			fmt.Println("Error accepting: ", e)
			continue
		}

		go HandleConnect(this, newConn)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <ms addr><r>\n", os.Args[0])
		os.Exit(-1)
	}

	// setup the kv service
	kvService := &KeyValService{
		NodeList:        make(map[string]net.Conn),
		KeyLocations:    make(map[string]map[string]struct{}), // eww
		UnavailableKeys: make(map[string]struct{}),
	}

	// get arguments
	msAddr := os.Args[1]

	// Listening to Clients
	go ListenConnections(kvService, msAddr)
}
