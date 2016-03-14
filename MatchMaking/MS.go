package main

import (
	//"encoding/json"
	"fmt"
	//"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const sessionDelay time.Duration = 250 * time.Millisecond

/////////// Debugging Helper

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

// check if a fatal error has ocurred
func FatalError(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(-10)
	}
}

/////////// RPC connection

// When clients first connect to the MS server
type Node struct {
	Id string // Napon
	Ip string // ip addr of Napon
}

type GameArgs struct {
	nodeList []Node // List of peer a node should talk to
}

// Reply from client
type ValReply struct {
	Val string // value; depends on the call
}

// main context
type Context struct {
	NodeLock sync.RWMutex

	gameList map[int][]Node // note: a map[string]struct{} acts list a Set<String> from Java

	MessageId int // atomically incremented message id
	roomID    int // atomically incremented game room id
	roomLimit int
}

// Reset context for the next session
func (this *Context) endSession() {
	this.NodeLock.Lock()

	this.NodeLock.Unlock()
}

// Notify all cients in current session about other players in the same room
func (this *Context) notifyClient() {
	//keys := make([]int, 0, len(this.gameList))

	this.NodeLock.Lock()

	/* Get all keys of gameList
	for k := range this.gameList {
		keys = append(keys, k)
	}

	fmt.Println("Keys:", keys)

	/*
		for k := range keys {
			clients := this.gameList[k]

			for c := range clients {
				// Trigger startGame() in c
			}
		}
	*/

	// For each client in a game room, trigger startGame() in those clients

	this.NodeLock.Unlock()

}

func (this *Context) Join(node *Node, reply *ValReply) error {

	AddNode(this, node)

	return nil
}

/////////// Helper methods

// this is called when a node joins, it handles adding the node to lists
func AddNode(this *Context, node *Node) {
	DebugPrint(1, "New Client "+node.Id)
	fmt.Println(node)

	this.NodeLock.Lock()

	// Check if room exists
	_, ok := this.gameList[this.roomID]
	if !ok {
		this.gameList[this.roomID] = make([]Node, 0)
	}

	// Check if room is full
	if len(this.gameList[this.roomID]) >= this.roomLimit {
		this.roomID++
		this.gameList[this.roomID] = make([]Node, 0)
	}

	// Add this client to the gameRoom
	this.gameList[this.roomID] = append(this.gameList[this.roomID], *node)

	// Look at the current roomID & check corresponding room
	fmt.Println("number of players:", this.gameList)

	this.NodeLock.Unlock()
}

func main() {
	// go run MS.go :4421
	if len(os.Args) != 2 {
		fmt.Println("Not enough arguments")
		os.Exit(-1)
	}

	// setup the kv service
	context := &Context{
		MessageId: 0,
		roomID:    0,
		roomLimit: 2,
		gameList:  make(map[int][]Node),
	}

	/* Start the timer
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// Trigger startGame on client side and reset timer
				fmt.Println("Tick at", <-ticker.C)

				// Keep track of users whose gameRoom is not full

				// Reset timer and keep all clients

				context.notifyClient()
				context.endSession()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	*/

	// get arguments
	rpcAddr, e := net.ResolveTCPAddr("tcp", os.Args[1])
	FatalError(e)

	DebugPrint(1, "Starting MS server")
	rpc.Register(context)
	listener, e := net.Listen("tcp", rpcAddr.String())
	FatalError(e)

	// start the rpc side
	rpc.Accept(listener)
	DebugPrint(1, "Exiting")
}
