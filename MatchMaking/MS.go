package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const sessionDelay time.Duration = 10 * time.Second
const RpcStartGame string = "NodeService.StartGame"
const leastPlayers int = 2

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

func CheckError(err error, n int) {
	if err != nil {
		fmt.Println(n, ": ", err)
		os.Exit(n)
	}
}

/////////// RPC connection

// When clients first connect to the MS server
type Node struct {
	Id string // Napon
	Ip string // ip addr of Napon
}

type GameArgs struct {
	NodeList []Node // List of peer a node should talk to
}

// Reply from client
type ValReply struct {
	Val string // value; depends on the call
}

// main context
type Context struct {
	NodeLock sync.RWMutex

	gameRoom  []Node // the only one game room contains all existing players
	roomID    int    // atomically incremented game room id
	roomLimit int
	gameTimer *time.Timer // timer until game start
}

// Notify all cients in current session about other players in the same room
func (this *Context) startGame() {
	this.NodeLock.Lock()
	for _, client := range this.gameRoom {
		c, e := rpc.Dial("tcp", client.Ip)
		CheckError(e, 1)
		var reply *ValReply = &ValReply{Val: ""}
		e = c.Call(RpcStartGame, &GameArgs{NodeList: this.gameRoom}, reply)
		CheckError(e, 6)
		c.Close()
	}

	// Clear the game room
	this.gameRoom = make([]Node, 0)

	// Reset the timer
	this.gameTimer.Reset(sessionDelay)

	this.NodeLock.Unlock()
}

// RPC join called by a client
func (this *Context) Join(node *Node, reply *ValReply) error {
	AddNode(this, node)

	// Check if the room is full
	if len(this.gameRoom) >= this.roomLimit {
		this.startGame()
	}
	return nil
}

// Perform certain operation every sessionDelay
func endSession(this *Context) {
	for t := range this.gameTimer.C {

		// at are at least 2 players in the room
		if len(this.gameRoom) >= leastPlayers {
			this.startGame()
			log.Println("ES: at least 2 players at ", t)
		} else {
			this.gameTimer.Reset(sessionDelay)
			log.Println("ES: not enough players to start. Clock reset")
			fmt.Println("gameRoom:", this.gameRoom, " len is ", len(this.gameRoom))
		}
	}
}

/////////// Helper methods

// this is called when a node joins, it handles adding the node to lists
func AddNode(this *Context, node *Node) {
	DebugPrint(1, "New Client "+node.Id)
	this.NodeLock.Lock()

	// Add this client to the gameRoom
	this.gameRoom = append(this.gameRoom, *node)

	fmt.Println("gameRoom:", this.gameRoom, " len is ", len(this.gameRoom))
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
		roomID:    0,
		roomLimit: 5,
		gameRoom:  make([]Node, 0),
		gameTimer: time.NewTimer(5 * time.Second),
	}

	go endSession(context)

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
