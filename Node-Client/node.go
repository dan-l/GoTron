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

// Message to be passed among nodes.
type Message struct {
	IsLeader          bool     // is this from the leader.
	IsDirectionChange bool     // is this a direction change update.
	DeadNodes         []string // id of dead nodes.
	Node              Node     // interval update struct.
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
var nodes []*Node         // All nodes in the game.
var myNode *Node          // My node.

// #LEADER specific.
var deadNodes []string // id of dead nodes found.

// Sync variables.
var waitGroup sync.WaitGroup // For internal processes.

// Game timers in milliseconds.
var intervalUpdateRate time.Duration
var tickRate time.Duration

var board [BOARD_SIZE][BOARD_SIZE]string
var directions map[string]string
var initialPosition map[string]*Pos
var lastCheckin map[string]time.Time

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

	initialPosition = map[string]*Pos{
		"p1": &Pos{1, 1},
		"p2": &Pos{8, 8},
		"p3": &Pos{8, 1},
		"p4": &Pos{1, 8},
		"p5": &Pos{4, 1},
		"p6": &Pos{5, 8},
	}

	nodes = make([]*Node, 0)
	lastCheckin = make(map[string]time.Time)
	deadNodes = make([]string, 0)
	tickRate = 500 * time.Millisecond
	intervalUpdateRate = 500 * time.Millisecond // TODO we said it's 100 in proposal?
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
	//client1 := Node{Id: "p1", Ip: "localhost:8767", CurrLoc: &Pos{1, 1}}
	//client2 := Node{Id: "p2", Ip: "localhost:8768", CurrLoc: &Pos{8, 8}}
	//client3 := Node{Id: "p3", Ip: "localhost:8769", CurrLoc: &Pos{8, 1}}
	//nodes = append(nodes, client1, client2, client3)

	// The above is commented out because we are now hooked up with the Matchmaking server.

	// Init everyone's location and find myself.
	for _, node := range nodes {
		node.Direction = directions[node.Id]
	}

	// find myself
	for _, node := range nodes {
		node.CurrLoc = initialPosition[node.Id]
		node.Direction = directions[node.Id]
		if node.Ip == nodeAddr {
			myNode = node
			nodeId = node.Id
		}
		lastCheckin[node.Id] = time.Now()
	}

	// ================================================= //

	isPlaying = true

	go listenUDPPacket()
	go intervalUpdate()
	go GameStateUpdate()
	go tickGame()
	go handleNodeFailure()
}

// Leade's Role: Leader notify other peers
func GameStateUpdate() {

}

// Each tick of the game
func tickGame() {
	if isPlaying == false {
		return
	}

	for {
		for i, node := range nodes {
			playerIndex := i + 1
			direction := node.Direction
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
				if node.Id == nodeId && isPlaying {
					gSO.Emit("playerDead")
					isPlaying = false
				}
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
	// TODO: This is a disgusting, terrible hack to allow the Node layer to
	//       broadcast state updates. We should replace this with something
	//       that's actually reasonable.
	if gSO != nil {
		gSO.Emit("gameStateUpdate", board)
	}
}

// Update peers with node's current location.
func intervalUpdate() {
	if isPlaying == false {
		return
	}

	for {
		var message *Message
		if isLeader() {
			message = &Message{IsLeader: true, DeadNodes: deadNodes, Node: *myNode}
		} else {
			message = &Message{Node: *myNode}
		}

		nodeJson, err := json.Marshal(message)
		//log.Println("Data to send: " + fmt.Sprintln(message))
		checkErr(err)
		sendPacketsToPeers(nodeJson)
		time.Sleep(intervalUpdateRate)
	}
}

func sendPacketsToPeers(payload []byte) {
	for _, node := range nodes {
		if node.Id != nodeId {
			data := send("Sending interval update to "+node.Id+" at ip "+node.Ip, payload)
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
		msg := receive("Received packet from "+addr.String(), buf, n)
		data := msg.Payload
		var message Message
		var node Node
		err = json.Unmarshal(data, &message)
		checkErr(err)
		node = message.Node

		log.Println("Received ", node)
		lastCheckin[node.Id] = time.Now()

		if message.IsLeader {
			log.Println("deadNodes are: ", message.DeadNodes)
			for _, n := range message.DeadNodes {
				removeNodeFromList(n)
			}

			// Construct history of each node based on the incoming message

		} else {

			// Store the hisotry from the leader
		}

		if message.IsDirectionChange {
			for _, n := range nodes {
				if n.Id == message.Node.Id {
					n.Direction = message.Node.Direction
				}
			}

		}

		if err != nil {
			log.Println("Error: ", err)
		}

		time.Sleep(400 * time.Millisecond)
	}
}

func notifyPeersDirChanged(direction string) {
	prevDirection := myNode.Direction

	// check if the direction change for node with the id
	if prevDirection != direction {
		log.Println("Direction for ", nodeId, " has changed from ",
			prevDirection, " to ", direction)
		myNode.Direction = direction
		directions[nodeId] = direction

		msg := &Message{IsDirectionChange: true, Node: *myNode}
		msgJson, err := json.Marshal(msg)
		checkErr(err)
		sendPacketsToPeers(msgJson)
	}
}

func isLeader() bool {
	return nodes[0].Id == nodeId
}

func hasExceededThreshold(nodeLastCheckin int64) bool {
	// TODO gotta check the math
	threshold := nodeLastCheckin + (700 * int64(time.Millisecond/time.Nanosecond))
	now := time.Now().UnixNano()
	//log.Println("Threshold ", threshold, "Now ", now)
	return threshold < now
}

func handleNodeFailure() {
	if isPlaying == false {
		return
	}

	// only for regular node
	// check if the time it last checked in exceed CHECKIN_INTERVAL
	for {
		if isLeader() {
			log.Println("Im a leader.")
			for _, node := range nodes {
				if node.Id != nodeId {
					if hasExceededThreshold(lastCheckin[node.Id].UnixNano()) {
						log.Println(node.Id, " HAS DIED")
						// TODO tell rest of nodes this node has died
						// --> leader should periodically send out active nodes in the system
						// --> so here we just have to remove it from the nodes list.
						deadNodes = append(deadNodes, node.Id)
						log.Println(len(deadNodes))
						removeNodeFromList(node.Id)
					}
				}
			}
		} else {
			log.Println("Im a node.")
			// Continually check if leader is alive.
			leaderId := nodes[0].Id
			if hasExceededThreshold(lastCheckin[leaderId].UnixNano()) {
				log.Println("LEADER ", leaderId, " HAS DIED.")
				removeNodeFromList(leaderId)
				// TODO: remove leader? or ask other peers first?
			}
		}
		time.Sleep(intervalUpdateRate)
	}
}

// LEADER: removes a dead node from the node list.
// TODO: Have to confirm if this works.
func removeNodeFromList(id string) {
	i := 0
	for i < len(nodes) {
		currentNode := nodes[i]
		if currentNode.Id == id {
			nodes = append(nodes[:i], nodes[i+1:]...)
		} else {
			i++
		}
	}
}

func leaderConflictResolution() {
	// as the referee of the game,
	// broadcast your game state for the current window to all peers
	// call sendUDPPacket
}

// Error checking. Exit program when error occurs.
func checkErr(err error) {
	if err != nil {
		log.Println("error:", err)
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
