package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Pos struct {
	X int
	Y int
}

// Peers
type Node struct {
	Id      string
	Ip      string
	currLoc Pos // internal use, so that it doesn't get exported during marshalling
}

const (
	BOARD_SIZE       int    = 10
	CHECKIN_INTERVAL int    = 200
	DIRECTION_UP     string = "U"
	DIRECTION_DOWN   string = "D"
	DIRECTION_LEFT   string = "L"
	DIRECTION_RIGHT  string = "R"
)

// Game variables.
var nodeId string         // Name of client.
var nodeAddr string       // IP of client.
var httpServerAddr string // HTTP Server IP.
var nodes []Node          // All nodes in the game.
var myNode Node           // My node.

// Sync variables.
var waitGroup sync.WaitGroup // For internal processes.

var board [10][10]string
var directions map[string]string
var playerMap map[string]string

func main() {
	if len(os.Args) < 4 {
		log.Println("usage: NodeClient [nodeAddr] [nodeRpcAddr] [msServerAddr] [httpServerAddr]")
		log.Println("[nodeAddr] the udp ip:port node is listening to")
		log.Println("[nodeRpcAddr] the rpc ip:port node is hosting for ms server")
		log.Println("[msServerAddr] the rpc ip:port of matchmaking server node is connecting to")
		log.Println("[httpServerAddr] the ip:port the http server is binded to ")
		os.Exit(1)
	}

	nodeAddr, nodeRpcAddr, msServerAddr, httpServerAddr = os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	log.Println(nodeAddr, nodeRpcAddr, msServerAddr, httpServerAddr)
	initLogging()

	// ============= FOR TESTING PURPOSES ============== //
	// Add myself.
	nodeId = "meeee"
	myNode = Node{Id: nodeId, Ip: nodeAddr, currLoc: Pos{1, 1}}

	// Add some enemies.
	client2 := Node{Id: "foo2", Ip: ":8768"}
	client3 := Node{Id: "foo3", Ip: ":8769"}

	nodes = append(nodes, myNode, client2, client3)
	// ================================================= //

	waitGroup.Add(3) // Add internal process.
	go msRpcServce()
	go httpServe()
	go intervalUpdate() // Internal update mechanism.
	waitGroup.Wait()    // Wait until processes are done.
}

// Initialize variables.
func init() {
	board = [10][10]string{
		[10]string{"", "", "", "", "", "", "", "", "", ""},
		[10]string{"", "p1", "", "", "", "", "", "", "p3", ""},
		[10]string{"", "", "", "", "", "", "", "", "", ""},
		[10]string{"", "", "", "", "", "", "", "", "", ""},
		[10]string{"", "p5", "", "", "", "", "", "", "", ""},
		[10]string{"", "", "", "", "", "", "", "", "p6", ""},
		[10]string{"", "", "", "", "", "", "", "", "", ""},
		[10]string{"", "", "", "", "", "", "", "", "", ""},
		[10]string{"", "p4", "", "", "", "", "", "", "p2", ""},
		[10]string{"", "", "", "", "", "", "", "", "", ""},
	}

	directions = map[string]string{
		"p1": DIRECTION_RIGHT,
		"p2": DIRECTION_LEFT,
		"p3": DIRECTION_LEFT,
		"p4": DIRECTION_RIGHT,
		"p5": DIRECTION_RIGHT,
		"p6": DIRECTION_LEFT,
	}

	playerMap = map[string]string{
		"id1": "p1",
		"id2": "p2",
		"id3": "p3",
		"id4": "p4",
		"id5": "p5",
		"id6": "p6",
	}

	nodes = make([]Node, 0)
}

// Update peers with node's current location.
func intervalUpdate() {
	defer waitGroup.Done()
	for {
		currentLocationJSON, err := json.Marshal(myNode.currLoc)
		log.Println("Data to send: " + fmt.Sprintln(myNode.currLoc))
		checkErr(err)
		for _, node := range nodes {
			if node.Id != nodeId {
				log.Println("Sending interval update to " + node.Id + " at ip " + node.Ip)
				sendUDPPacket(node.Ip, currentLocationJSON)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// Send data to ip via UDP.
func sendUDPPacket(ip string, data []byte) {
	localAddr, err := net.ResolveUDPAddr("udp", nodeAddr)
	checkErr(err)
	ipAddr, err := net.ResolveUDPAddr("udp", ip)
	checkErr(err)
	udpConn, err := net.DialUDP("udp", localAddr, ipAddr)
	checkErr(err)

	defer udpConn.Close()

	_, err = udpConn.Write(data)
	checkErr(err)
}

func handleNodeFailure() {
	// only for regular node
	// check if the time it last checked in exceed CHECKIN_INTERVAL
	// mark the alive property on node object
}

func leaderConflictResolution() {
	// as the referee of the game,
	// broadcast your game state for the current window to all peers
	// call sendUDPPacket
}

// Error checking. Exit program when error occurs.
func checkErr(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
