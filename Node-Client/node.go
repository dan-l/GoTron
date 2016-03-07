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
	id string
	ip string
}

const (
	boardSize int = 10
)

// Game variables.
var nodeId string         // Name of client.
var nodeAddr string       // IP of client.
var msServerAddr string   // Matchmaking server IP.
var httpServerAddr string // HTTP Server IP.
var gameState GameState   // Overall GameState of the client.
var nodes []Node          // All nodes in the game.

// Sync variables.
var waitGroup sync.WaitGroup // For internal processes.
var done chan int            // For external processes.

func main() {
	if len(os.Args) < 5 {
		log.Println("usage: node.go [nodeid] [nodeAddr] [msServerAddr] [httpServerAddr]")
		log.Println("[nodeid] unique id for each player")
		log.Println("[nodeAddr] the rpc ip:port the node is listening to")
		log.Println("[msServerAddr] the rpc ip:port of matchmaking server is listening to")
		log.Println("[httpServerAddr] the ip:port the http server is binded to ")
		os.Exit(1)
	}

	nodeId, nodeAddr, msServerAddr, httpServerAddr = os.Args[1], os.Args[2], os.Args[3], os.Args[4]
	init() // Initialize variables.

	log.Println(nodeId, nodeAddr, msServerAddr, httpServerAddr)
	initLogging()

	httpMsg := make(chan string)
	rpcMsg := make(chan string)
	go msRpcServce(rpcMsg, msServerAddr)
	go httpServe(httpMsg)
	<-done
	<-done

	waitGroup.Add(1)    // Add an internal process.
	go intervalUpdate() // Internal update mechanism.

	// Wait until processes are done.
	waitGroup.Wait()
}

// Initialize variables.
func init() {
	done = make(chan int, 2)
	board := make([][]string, boardSize)
	for i := range board {
		board[i] = make([]string, boardSize)
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
	for {
		currentLocationJSON, err := json.Marshal(gameState.currLocs[clientName])
		checkErr(err)
		for i, node := range nodes {
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

// Error checking. Exit program when error occurs.
func checkErr(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
