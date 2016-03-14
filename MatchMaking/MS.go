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

var EmptyStruct struct{}

// main context
type KeyValService struct {
	NodeLock sync.RWMutex

	NodeList map[string]net.Conn
	gameList map[int32]map[string]struct{} // note: a map[string]struct{} acts list a Set<String> from Java

	MessageId int32 // atomically incremented message id
	roomID    int32 // atomically incremented game room id
	roomLimit int32
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
const DebugLevel int = 4

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

// this is called when a node joins, it handles adding the node to lists
func AddNode(this *KeyValService, hello *HelloMessage, conn net.Conn) {
	DebugPrint(1, "New Client"+hello.Id)
	//keys := make(map[string]struct{})

	// Look at the current roomID & check corresponding room
	fmt.Println("number of players:", this.gameList[this.roomLimit])

	this.NodeLock.Lock()
	this.NodeList[hello.Id] = conn
	//this.gameList[this.roomID] =

	fmt.Println("NodeList:", this.NodeList)

	this.NodeLock.Unlock()
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

	// store data locally
	AddNode(this, &hello, conn)

	DebugPrint(1, "SUCCESS")
}

func main() {
	// go run MS.go :4421
	if len(os.Args) != 2 {
		fmt.Println("Not enough arguments")
		os.Exit(-1)
	}

	// setup the kv service
	kvService := &KeyValService{
		NodeList:  make(map[string]net.Conn),
		MessageId: 0,
		roomID:    0,
		roomLimit: 5,
		gameList:  make(map[int32]map[string]struct{}), // eww
	}

	// get arguments
	msAddr, e := net.ResolveTCPAddr("tcp", os.Args[1])
	FatalError(e)

	// Listening to Clients
	DebugPrint(1, "Starting MS server")
	conn, e := net.Listen("tcp", msAddr.String())
	defer conn.Close()
	FatalError(e)

	for {
		fmt.Println("Reading once from TCP connection")
		newConn, e := conn.Accept()
		if e != nil {
			fmt.Println("Error accepting: ", e)
			continue
		}
		go HandleConnect(kvService, newConn)
	}
	DebugPrint(1, "Exiting")
}
