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
	Id        string
	Ip        string // udp port this node is listening to
	CurrLoc   *Pos
	Direction string
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

	startGame() // TODO Testing purposes only, should be called in rpc start game

	waitGroup.Add(2) // Add internal process.
	go msRpcServce()
	go httpServe()

	waitGroup.Wait() // Wait until processes are done.
}

// Initialize variables.
func init() {
	board = [BOARD_SIZE][BOARD_SIZE]string{
		[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
		[BOARD_SIZE]string{"", "p1", "", "", "", "", "", "", "p3", ""},
		[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
		[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
		[BOARD_SIZE]string{"", "p5", "", "", "", "", "", "", "", ""},
		[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "p6", ""},
		[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
		[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
		[BOARD_SIZE]string{"", "p4", "", "", "", "", "", "", "p2", ""},
		[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
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

func intMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func intMin(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func startGame() {
	// ============= FOR TESTING PURPOSES ============== //
	// Add some enemies. TODO: CurrLoc for peers should be initialized elsewhere.
	// Hardcoded list of clients
	// NOTE: Do not use addresses that lack an IP/hostname such as ":8767".
	//       It breaks running the program on Windows.
	client1 := Node{Id: "p1", Ip: "localhost:8767", CurrLoc: &Pos{1, 1}}
	client2 := Node{Id: "p2", Ip: "localhost:8768", CurrLoc: &Pos{8, 8}}
	client3 := Node{Id: "p3", Ip: "localhost:8769", CurrLoc: &Pos{8, 1}}

	nodes = append(nodes, client1, client2, client3)

	// find myself
	for _, node := range nodes {
		if node.Ip == nodeAddr {
			myNode = node
			nodeId = node.Id
		}
	}

	// ================================================= //

	isPlaying = true

	go listenUDPPacket()
	go intervalUpdate()
	// go tickGame()
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
			direction := directions[node.Id]
			x := node.CurrLoc.X
			y := node.CurrLoc.Y
			new_x := node.CurrLoc.X
			new_y := node.CurrLoc.Y

			// Path prediction
			board[y][x] = "t" + strconv.Itoa(playerIndex) // Change position to be a trail.
			switch direction {
			case DIRECTION_UP:
				new_y = intMax(0, y-1)
			case DIRECTION_DOWN:
				new_y = intMin(BOARD_SIZE-1, y+1)
			case DIRECTION_LEFT:
				new_x = intMax(0, x-1)
			case DIRECTION_RIGHT:
				new_x = intMin(BOARD_SIZE-1, x+1)
			}

			if nodeHasCollided(x, y, new_x, new_y) {
				log.Println("NODE " + node.Id + " IS DEAD")
				// We don't update the position to a new value
				board[y][x] = "d" + strconv.Itoa(playerIndex) // Dead node
			} else {
				// Update player's new position.
				board[new_y][new_x] = "p" + strconv.Itoa(playerIndex)
				node.CurrLoc.X = new_x
				node.CurrLoc.Y = new_y
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
		nodeJson, err := json.Marshal(myNode)
		log.Println("Data to send: " + fmt.Sprintln(myNode))
		checkErr(err)
		sendPacketsToPeers(nodeJson)
		time.Sleep(intervalUpdateRate)
	}
}

func sendPacketsToPeers(data []byte) {
	for _, node := range nodes {
		if node.Id != nodeId {
			log.Println("Sending interval update to " + node.Id + " at ip " + node.Ip)
			sendUDPPacket(node.Ip, data)
		}
	}
}

// Send data to ip via UDP.
func sendUDPPacket(ip string, data []byte) {
	// TODO a random port is picked since
	// we can't listen and read at the same time
	udpConn, err := net.Dial("udp", ip)
	checkErr(err)
	defer udpConn.Close()

	_, err = udpConn.Write(data)
	checkErr(err)
}

func listenUDPPacket() {
	localAddr, err := net.ResolveUDPAddr("udp", nodeAddr)
	checkErr(err)
	udpConn, err := net.ListenUDP("udp", localAddr)
	checkErr(err)
	defer udpConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := udpConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func findCurrLoc() *Pos {
	for i, _ := range board {
		for j, p := range board[i] {
			if p == nodeId {
				return &Pos{i, j}
			}
		}
	}
	return nil
}

func notifyPeersDirChanged(direction string) {
	log.Println("Check if dir changed")
	// check if the direction change for node with the id
	if directions[nodeId] != direction {
		log.Println("Direction for ", nodeId, " has changed from ",
			directions[nodeId], " to ", direction)
		myNode.Direction = direction
		currLoc := findCurrLoc()
		if currLoc != nil {
			myNode.CurrLoc = currLoc
		} else {
			log.Fatal("IM LOSTTTTTT")
		}

		nodeJson, err := json.Marshal(myNode)
		log.Println("Data to send: " + fmt.Sprintln(myNode))
		checkErr(err)
		sendPacketsToPeers(nodeJson)
	}
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
