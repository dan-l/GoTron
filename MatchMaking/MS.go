package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

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

// The program should exit if this gives error
func FatalError(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(-10)
	}
}

// Help debug the location
func CheckError(err error, n int) {
	if err != nil {
		fmt.Println(n, ": ", err)
		os.Exit(n)
	}
}

/////////// RPC connection

// When clients first connect to the MS server
type Node struct {
	Id string // [p1 to p6]
	Ip string // ip to send to each player
}

type NodeJoin struct {
	RpcIp string // The one MS has to dial at start Game
	Ip    string // ip to send to each player
}

type GameArgs struct {
	NodeList []*Node // List of peer a node should talk to
}

// Reply from client
type ValReply struct {
	Val string // value; depends on the call
}

// main context
type Context struct {
	NodeLock sync.RWMutex

	nodeList  []*NodeJoin
	gameRoom  []*Node // the only one game room contains all existing players
	roomID    int     // atomically incremented game room id
	roomLimit int
	gameTimer *time.Timer // timer until game start

}

// Assign id to each client
func (this *Context) assignID() {
	this.NodeLock.Lock()
	for index, client := range this.gameRoom {
		client.Id = "p" + strconv.Itoa(index+1)
	}
	this.NodeLock.Unlock()
	fmt.Println("Finish assigningID:", this.gameRoom)
}

// Notify all cients in current session about other players in the same room
func (this *Context) startGame() {
	this.NodeLock.Lock()
	for _, client := range this.nodeList {
		c, e := rpc.Dial("tcp", client.RpcIp)
		CheckError(e, 1)
		var reply *ValReply = &ValReply{Val: ""}
		e = c.Call(RpcStartGame, &GameArgs{NodeList: this.gameRoom}, reply)
		CheckError(e, 6)
		c.Close()
	}

	// Clear the game room &
	this.gameRoom = make([]*Node, 0)

	// Reset the timer
	this.gameTimer.Reset(sessionDelay)

	this.NodeLock.Unlock()
}

// RPC join called by a client
func (this *Context) Join(node *NodeJoin, reply *ValReply) error {
	AddNode(this, node)

	// Check if the room is full
	if len(this.gameRoom) >= this.roomLimit {
		this.assignID()
		this.startGame()
	}

	return nil
}

// Perform certain operation every sessionDelay
func endSession(this *Context) {
	for t := range this.gameTimer.C {
		// At are at least 2 players in the room
		if len(this.gameRoom) >= leastPlayers {
			this.assignID()
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

// this is called when a node disconnects
func deleteNode(ctx *Context, rpcip string) {
	ctx.NodeLock.Lock()
	DebugPrint(1, "Lost Node:"+rpcip)

	ctx.NodeLock.Unlock()
}

// this is called when a node joins, it handles adding the node to lists
func AddNode(ctx *Context, nodeJoin *NodeJoin) {
	fmt.Println("new node:", nodeJoin)
	ctx.NodeLock.Lock()

	// Add this client to the gameRoom
	ctx.nodeList = append(ctx.nodeList, nodeJoin)

	node := &Node{Ip: nodeJoin.Ip}
	ctx.gameRoom = append(ctx.gameRoom, node)

	fmt.Println("gameRoom:", ctx.gameRoom, " len is ", len(ctx.gameRoom))
	ctx.NodeLock.Unlock()
}

// Listen and serve request from client
func listenToClient(ctx *Context, rpcAddr string) {
	waitGroup.Done()
	for {
		rpc.Register(ctx)
		listener, e := net.Listen("tcp", rpcAddr)
		FatalError(e)
		fmt.Println("LISTENING")

		for {
			connection, e := listener.Accept()
			FatalError(e)
			defer connection.Close()

			// Handle one connection at a time
			go rpc.ServeConn(connection)
		}
		time.Sleep(time.Millisecond * 100)
	}
}

// Global variables
var waitGroup sync.WaitGroup // Wait group
const sessionDelay time.Duration = 10 * time.Second
const RpcStartGame string = "NodeService.StartGame"
const leastPlayers int = 2

func main() {
	// go run MS.go :4421
	if len(os.Args) != 2 {
		fmt.Println("Not enough arguments")
		os.Exit(-1)
	}

	// setup the kv service
	context := &Context{
		nodeList:  make([]*NodeJoin, 0),
		roomID:    0,
		roomLimit: 5,
		gameRoom:  make([]*Node, 0),
		gameTimer: time.NewTimer(5 * time.Second),
	}

	// get arguments
	rpcAddr, e := net.ResolveTCPAddr("tcp", os.Args[1])
	FatalError(e)
	DebugPrint(1, "Starting MS server")

	waitGroup.Add(2)

	go endSession(context)
	go listenToClient(context, rpcAddr.String())

	// Wait until processes are done.
	waitGroup.Wait()
}
