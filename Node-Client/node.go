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

type GameState struct {
	board    Board
	currLocs map[string]Location // where the players are now
	stat     GameStat
}

type GameStat struct {
	scores map[string]int
}

// Board is represented by string with value of
// id if it's occupied by corresponding player or "" if unoccupied
type Board [][]string

type Location struct {
	x   int
	y   int
	dir string // 'U', 'D', 'L', 'R'
}

// Peers
type Node struct {
	id    string
	ip    string
}

const (
	BOARD_SIZE       int = 10
	CHECKIN_INTERVAL int = 200
)

// Game variables.
var nodeId string         // Name of client.
var nodeAddr string       // IP of client.
var httpServerAddr string // HTTP Server IP.
var gameState GameState   // Overall GameState of the client.
var nodes []Node          // All nodes in the game.

// Sync variables.
var waitGroup sync.WaitGroup // For internal processes.

var BOARD = [10][10]string{
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

func main() {
	if len(os.Args) < 5 {
		log.Println("usage: NodeClient [nodeid] [nodeAddr] [nodeRpcAddr] [msServerAddr] [httpServerAddr]")
		log.Println("[nodeid] unique id for each player")
		log.Println("[nodeAddr] the udp ip:port node is listening to")
		log.Println("[nodeRpcAddr] the rpc ip:port node is hosting for ms server")
		log.Println("[msServerAddr] the rpc ip:port of matchmaking server node is connecting to")
		log.Println("[httpServerAddr] the ip:port the http server is binded to ")
		os.Exit(1)
	}

	nodeId, nodeAddr, nodeRpcAddr, msServerAddr, httpServerAddr = os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5]

	log.Println(nodeId, nodeAddr, nodeRpcAddr, msServerAddr, httpServerAddr)
	initLogging()

	waitGroup.Add(3) // Add internal process.
	go msRpcServce()
	go httpServe()
	go intervalUpdate() // Internal update mechanism.
	waitGroup.Wait()    // Wait until processes are done.
}

// Initialize variables.
func init() {
	board := make([][]string, BOARD_SIZE)
	for i := range board {
		board[i] = make([]string, BOARD_SIZE)
	}
	currLocs := make(map[string]Location)
	scores := make(map[string]int)
	gameStat := GameStat{scores}
	gameState = GameState{board, currLocs, gameStat}
	nodes = make([]Node, 0)
	log.Println(gameState)
}

// Update peers with node's current location.
func intervalUpdate() {
	defer waitGroup.Done()
	for {
		currentLocationJSON, err := json.Marshal(gameState.currLocs[nodeId])
		checkErr(err)
		for _, node := range nodes {
			if node.id != nodeId {
				sendUDPPacket(node.ip, currentLocationJSON)
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
