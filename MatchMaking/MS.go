package main

import (
	"fmt"
	//"log"
	"net"
	//"net/rpc"
	"encoding/json"
	//"github.com/deckarep/golang-set"
	"os"
	"sync"
)

var EmptyStruct struct{}

// main context
type KeyValService struct {
	NodeLock sync.RWMutex

	NodeList map[string]net.Conn
	gameList map[int]map[string]struct{} // note: a map[string]struct{} acts list a Set<String> from Java

	MessageId int // atomically incremented message id
	roomID    int // atomically incremented game room id
	roomLimit int
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

	this.NodeLock.Lock()

	// Check if room exists
	_, ok := this.gameList[this.roomID]
	if !ok {
		this.gameList[this.roomID] = make(map[string]struct{})
	}

	// Check if room is full
	if len(this.gameList[this.roomID]) >= this.roomLimit {
		this.roomID++
		this.gameList[this.roomID] = make(map[string]struct{})
	}

	// Add this client to the gameRoom
	this.gameList[this.roomID][hello.Id] = EmptyStruct

	// Look at the current roomID & check corresponding room
	fmt.Println("number of players:", this.gameList)

	this.NodeList[hello.Id] = conn

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
		roomLimit: 2,
		gameList:  make(map[int]map[string]struct{}), // eww
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
