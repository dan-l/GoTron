package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	//"sort"
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

	connections map[string]*rpc.Client // Client's IPaddr : connection
	nodeList    map[string]*Node       // map rpcIP to a node object
	gameRoom    []*Node                // the only one game room contains all existing players
	clientNum   int                    // the order of incoming clients
	roomLimit   int
	gameTimer   *time.Timer // timer until game start

}

// Construct a game room from nodeList
func (this *Context) makeGameRoom() {
	fmt.Println("Making a Game room")
	for _, node := range this.nodeList {
		this.gameRoom = append(this.gameRoom, node)
	}

}

// Assign id to each client
func (this *Context) assignID() {
	fmt.Println("Assigning IDs")
	for index, client := range this.gameRoom {
		client.Id = "p" + strconv.Itoa(index+1)
	}
}

// Notify all cients in current session about other players in the same room
func (this *Context) startGame() {
	fmt.Println("Connection Number:", len(this.connections))
	for key, _ := range this.nodeList {
		//fmt.Println("Trying to dial", key)
		//fmt.Println("dialed", key)
		var reply *ValReply = &ValReply{Val: ""}
		//fmt.Println("calling startgame with connections = ", this.connections)
		//fmt.Println("c is:", this.connections[key])
		e := this.connections[key].Call(RpcStartGame, &GameArgs{NodeList: this.gameRoom}, reply)
		if e != nil {
			fmt.Println("Failed to start", key)
		}
		//fmt.Println("startd")
	}

	// Clear the game room, nodelist, and connections
	this.gameRoom = make([]*Node, 0)
	this.nodeList = make(map[string]*Node)
	this.connections = make(map[string]*rpc.Client)

	// Reset the timer
	this.gameTimer.Reset(sessionDelay)
	//fmt.Println("DONE StartGame")
}

// Update NodeList and Connection based on disconnected clients
func (this *Context) checkConn() {
	this.NodeLock.Lock()

	// client in the connections -> no need to dial, just call
	// client NOT in the connections -> dial first and call
	for ClientIp, _ := range this.nodeList {
		_, exist := this.connections[ClientIp]
		if exist {
			var reply *ValReply = &ValReply{Val: ""}
			e := this.connections[ClientIp].Call(RpcMessage, &GameArgs{NodeList: this.gameRoom}, reply)
			if e != nil {
				fmt.Println(e)
				fmt.Println("Deleting disconnected node ", ClientIp)
				delete(this.nodeList, ClientIp)
				continue
			} else {
				// Update connection for each client
				fmt.Println("client: ", ClientIp, " is good.")
			}
		} else {
			c, e := rpc.Dial("tcp", ClientIp)
			if e != nil {
				fmt.Println(e)
				fmt.Println("Deleting disconnected node ", ClientIp)
				delete(this.nodeList, ClientIp)
				continue
			} else {
				// Update connection for each client
				fmt.Println("client: ", ClientIp, " is good.")
				//fmt.Println("c is", c)
				this.connections[ClientIp] = c
			}
		}

	}
	this.NodeLock.Unlock()
}

// RPC join called by a client
func (this *Context) Join(nodeJoin *NodeJoin, reply *ValReply) error {
	AddNode(this, nodeJoin)

	this.checkConn() // Update NodeList and Connections

	// Check if the room is full
	if len(this.nodeList) >= this.roomLimit {
		this.NodeLock.Lock()
		fmt.Println("Join: Starting Game")
		this.makeGameRoom() // NodeList -> GameROOM
		this.assignID()     // Assign corresponding ID to gameROOM
		this.startGame()
		this.NodeLock.Unlock()
	}
	return nil
}

// Perform certain operation every sessionDelay
func endSession(this *Context) {
	defer waitGroup.Done()
	for _ = range this.gameTimer.C {

		this.checkConn() // Update NodeList and Connections

		// At are at least 2 players in the room
		if len(this.nodeList) >= leastPlayers {
			this.NodeLock.Lock()
			fmt.Println("ES: Starting Game")
			this.makeGameRoom()
			this.assignID()
			this.startGame()
			log.Println("ES: Done Start Game")
			this.NodeLock.Unlock()
		} else {
			this.gameTimer.Reset(sessionDelay)
			fmt.Println("ES:", len(this.nodeList), "players")
			//log.Println("ES:", len(this.nodeList), "players currently.")
		}

	}
}

/////////// Helper methods

// this is called when a node joins, it handles adding the node to lists
func AddNode(ctx *Context, nodeJoin *NodeJoin) {
	ctx.NodeLock.Lock()
	fmt.Println("AD: new node:", nodeJoin)
	// Add this client to the gameRoom & NodeList
	node := &Node{Ip: nodeJoin.Ip}
	ctx.nodeList[nodeJoin.RpcIp] = node

	fmt.Println("AD: NodeList:", ctx.nodeList, ". Numb:", len(ctx.nodeList), "players.")
	ctx.NodeLock.Unlock()
}

// Listen and serve request from client
func listenToClient(ctx *Context, rpcAddr string) {
	defer waitGroup.Done()
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
const sessionDelay time.Duration = 5 * time.Second
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
		connections: make(map[string]*rpc.Client),
		nodeList:    make(map[string]*Node),
		clientNum:   0,
		roomLimit:   3,
		gameRoom:    make([]*Node, 0),
		gameTimer:   time.NewTimer(5 * time.Second),
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
