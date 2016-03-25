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
	}
}

/////////// RPC connection

// Object to be sent back to the client
type Node struct {
	Id string // [p1 to p6]
	Ip string // ip to send to each player
}

// Object received from the clients at the start
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

	nodeList  map[string]*Node // map rpcIP to a node object
	gameRoom  []*Node          // the only one game room contains all existing players
	roomID    int              // atomically incremented game room id
	roomLimit int
	gameTimer *time.Timer // timer until game start

}

// Assign id to each client
func (this *Context) assignID() {
	for index, client := range this.gameRoom {
		client.Id = "p" + strconv.Itoa(index+1)
	}
	fmt.Println("Finish assigningID:", this.gameRoom)
}

// Notify all cients in current session about other players in the same room
func (this *Context) startGame() {
	fmt.Println("Starting Game")
	for key, _ := range this.nodeList {
		fmt.Println("key in sg:", key)
		c, e := rpc.Dial("tcp", key)
		CheckError(e, 1)

		fmt.Println("Dial in sg:", this.gameRoom)
		var reply *ValReply = &ValReply{Val: ""}
		e = c.Call(RpcStartGame, &GameArgs{NodeList: this.gameRoom}, reply)
		CheckError(e, 6)

		fmt.Println("Call in sg")
		c.Close()

		fmt.Println("CLose in sg")
	}

	// Clear the game room &
	this.gameRoom = make([]*Node, 0)
	this.nodeList = make(map[string]*Node)

	// Reset the timer
	this.gameTimer.Reset(sessionDelay)
	fmt.Println("DONE StartGame")
}

// Construct a game room from nodeList
func (this *Context) makeGameRoom() {
	for _, node := range this.nodeList {
		this.gameRoom = append(this.gameRoom, node)
	}
}

// Ping clients to check if the connection is still good
func (this *Context) checkConn() {
	//this.NodeLock.Lock()
	for ClientIp, _ := range this.nodeList {
		c, e := rpc.Dial("tcp", ClientIp)
		if e != nil {
			fmt.Println(e)
			fmt.Println("1: Client:", ClientIp, " disconnect")
			deleteNode(this, ClientIp)
			continue
		}
		c.Close()
	}

	//this.NodeLock.Unlock()
	fmt.Println("Updated NodeList:", this.nodeList)
}

// RPC join called by a client
func (this *Context) Join(nodeJoin *NodeJoin, reply *ValReply) error {
	AddNode(this, nodeJoin)

	// Check all clients' connections
	//this.checkConn()

	// Check if the room is full
	if len(this.nodeList) >= this.roomLimit {
		this.NodeLock.Lock()
		this.makeGameRoom()
		this.assignID()
		this.startGame()
		this.NodeLock.Unlock()
	}

	return nil
}

// Perform certain operation every sessionDelay
func endSession(this *Context) {
	// defer waitGroup.Done()
	for t := range this.gameTimer.C {
		// Check connection first:
		this.checkConn()

		// At are at least 2 players in the room
		if len(this.nodeList) >= leastPlayers {
			this.NodeLock.Lock()
			this.makeGameRoom()
			this.assignID()
			this.startGame()
			log.Println("ES: at least 2 players at ", t)
			this.NodeLock.Unlock()
		} else {
			this.gameTimer.Reset(sessionDelay)
			log.Println("ES: NP")
			fmt.Println("nodeList:", this.nodeList, " len is ", len(this.nodeList))
		}
	}
}

/////////// Helper methods

// this is called when a node disconnects
func deleteNode(ctx *Context, rpcip string) {
	DebugPrint(1, "Lost Node:"+rpcip)
	delete(ctx.nodeList, rpcip)
}

// this is called when a node joins, it handles adding the node to lists
func AddNode(ctx *Context, nodeJoin *NodeJoin) {
	ctx.NodeLock.Lock()
	fmt.Println("new node:", nodeJoin)
	// Add this client to the gameRoom & NodeList
	node := &Node{Ip: nodeJoin.Ip}
	ctx.nodeList[nodeJoin.RpcIp] = node

	fmt.Println("NodeList:", ctx.nodeList, " len is ", len(ctx.nodeList))
	ctx.NodeLock.Unlock()
}

// Listen and serve request from client
func listenToClient(ctx *Context, rpcAddr string) {
	// defer waitGroup.Done()
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
const sessionDelay time.Duration = 6 * time.Second
const RpcStartGame string = "NodeService.StartGame"
const RpcMessage string = "NodeService.Message"
const leastPlayers int = 2

func main() {
	// go run MS.go :4421
	if len(os.Args) != 2 {
		fmt.Println("Not enough arguments")
		os.Exit(-1)
	}

	// setup the kv service
	context := &Context{
		nodeList:  make(map[string]*Node),
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

	go endSession(context) // Timer
	go listenToClient(context, rpcAddr.String())

	// Wait until processes are done.
	waitGroup.Wait()
}
