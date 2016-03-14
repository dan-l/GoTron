package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

var EmptyStruct struct{}

const sessionDelay time.Duration = 250 * time.Millisecond

// main context
type Context struct {
	NodeLock sync.RWMutex

	NodeList map[string]net.Conn
	gameList map[int]map[string]struct{} // note: a map[string]struct{} acts list a Set<String> from Java

	MessageId int // atomically incremented message id
	roomID    int // atomically incremented game room id
	roomLimit int
}

// Reset context for the next session
func (this *Context) clearContext() {
	this.NodeLock.Lock()
	this.NodeList = make(map[string]net.Conn)
	this.gameList = make(map[int]map[string]struct{})
	this.roomID = 0
	this.MessageId = 0
	this.NodeLock.Unlock()
}

// Notify all cients in current session about other players in the same room
func (this *Context) notifyClient() {
	this.NodeLock.Lock()
	// Get all keys of gameList

	// For each client in a game room, trigger startGame() in those clients

	this.NodeLock.Unlock()
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

// When clients first connect to the MS server
type HelloMessage struct {
	Id string
}

// this is called when a node joins, it handles adding the node to lists
func AddNode(this *Context, hello *HelloMessage, conn net.Conn) {
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
func HandleConnect(this *Context, conn net.Conn) {
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

}

func main() {
	// go run MS.go :4421
	if len(os.Args) != 2 {
		fmt.Println("Not enough arguments")
		os.Exit(-1)
	}

	// setup the kv service
	context := &Context{
		NodeList:  make(map[string]net.Conn),
		MessageId: 0,
		roomID:    0,
		roomLimit: 2,
		gameList:  make(map[int]map[string]struct{}),
	}

	// Start the timer
	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// Trigger startGame on client side and reset timer
				fmt.Println("Tick at", <-ticker.C)
				context.clearContext()
				context.notifyClient()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

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
		go HandleConnect(context, newConn)
	}
	DebugPrint(1, "Exiting")
}
