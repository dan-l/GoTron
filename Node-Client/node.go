package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
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
	currLoc *Pos // internal use, so that it doesn't get exported during marshalling
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
var isPlaying bool        // Is the game in session.
var nodeId string         // Name of client.
var nodeAddr string       // IP of client.
var httpServerAddr string // HTTP Server IP.
var nodes []Node          // All nodes in the game.
var myNode Node           // My node.

// Sync variables.
var waitGroup sync.WaitGroup // For internal processes.

// Game timers in milliseconds.
var intervalUpdateRate time.Duration
var tickRate time.Duration

var board [10][10]string
var directions map[string]string
var playerMap map[string]string

func main() {
	if len(os.Args) != 5 {
		log.Println("usage: NodeClient [nodeAddr] [nodeRpcAddr] [msServerAddr] [httpServerAddr]")
		log.Println("[nodeAddr] the udp ip:port node is listening to")
		log.Println("[nodeRpcAddr] the rpc ip:port node is hosting for ms server")
		log.Println("[msServerAddr] the rpc ip:port of matchmaking server node is connecting to")
		log.Println("[httpServerAddr] the ip:port the http server is binded to ")
		os.Exit(1)
	}

	nodeAddr, nodeRpcAddr, msServerAddr = os.Args[1], os.Args[2], os.Args[3]

	httpServerTcpAddr, err := net.ResolveTCPAddr("tcp", os.Args[4])
	checkErr(err)
	httpServerAddr = httpServerTcpAddr.String()

	log.Println(nodeAddr, nodeRpcAddr, msServerAddr, httpServerAddr)
	initLogging()

	// ============= FOR TESTING PURPOSES ============== //
	// Add myself.
	nodeId = "meeee"
	myNode = Node{Id: nodeId, Ip: nodeAddr, currLoc: &Pos{1, 1}}

	// Add some enemies. TODO: currLoc for peers should be initialized elsewhere.
	client2 := Node{Id: "foo2", Ip: "localhost:8768", currLoc: &Pos{8, 8}}
	client3 := Node{Id: "foo3", Ip: "localhost:8769", currLoc: &Pos{8, 1}}

	nodes = append(nodes, myNode, client2, client3)

	playerMap = map[string]string{
		nodeId:     "p1",
		client2.Id: "p2",
		client3.Id: "p3",
	}

	isPlaying = true
	// ================================================= //

	waitGroup.Add(4) // Add internal process.
	go msRpcServce()
	go httpServe()
	go intervalUpdate() // Internal update mechanism.
	go tickGame()       // Each tick of the game.
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

	nodes = make([]Node, 0)

	tickRate = 500 * time.Millisecond
	intervalUpdateRate = 500 * time.Millisecond
}

// Each tick of the game.
func tickGame() {
	defer waitGroup.Done()

	if isPlaying == false {
		return
	}

	for {
		for i, node := range nodes {
			playerIndex := i + 1
			direction := directions[playerMap[node.Id]]
			x := node.currLoc.X
			y := node.currLoc.Y
			new_x := node.currLoc.X
			new_y := node.currLoc.Y

			// Path prediction
			board[y][x] = "t" + strconv.Itoa(playerIndex) // Change position to be a trail.
			switch direction {
			case DIRECTION_UP:
				new_y = y - 1
			case DIRECTION_DOWN:
				new_y = y + 1
			case DIRECTION_LEFT:
				new_x = x - 1
			case DIRECTION_RIGHT:
				new_x = x + 1
			}

			if nodeHasCollided(x, y, new_x, new_y) {
				log.Println("NODE " + node.Id + " IS DEAD")
				// We don't update the position to a new value
				board[y][x] = "d" + strconv.Itoa(playerIndex) // Dead node
			} else {
				// Update player's new position.
				board[new_y][new_x] = "p" + strconv.Itoa(playerIndex)
				node.currLoc.X = new_x
				node.currLoc.Y = new_y
			}
		}
		renderGame()
		time.Sleep(tickRate)
	}
}

// Check if a node has collided into a trail, wall, or another node.
func nodeHasCollided(oldX int, oldY int, newX int, newY int) bool {
	// Wall boundaries.
	if newX < 0 || newY < 0 || newX > BOARD_SIZE || newY > BOARD_SIZE {
		return true
	}
	// Collision with another player or trail.
	if board[newY][newX] != "" {
		return true
	}
	return false
}

// Renders the game.
func renderGame() {
	printBoard()
}

// Update peers with node's current location.
func intervalUpdate() {
	defer waitGroup.Done()

	if isPlaying == false {
		return
	}

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
		time.Sleep(intervalUpdateRate)
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

// For debugging
func printBoard() {
	for r, _ := range board {
		fmt.Print("[")
		for _, item := range board[r] {
			if item == "" {
				fmt.Print("__" + " ")
			} else {
				fmt.Print(item + " ")
			}
		}
		fmt.Print("]\n")
	}
}
