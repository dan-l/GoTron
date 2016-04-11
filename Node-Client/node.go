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
	IsAlive   bool
}

// Message to be passed among nodes.
type Message struct {
	IsLeader          bool                // is this from the leader.
	IsDirectionChange bool                // is this a direction change update.
	IsDeathReport     bool                // is this a death report.
	FailedNodes       []string            // id of disconnected nodes.
	Node              Node                // interval update struct node or dead node.
	GameHistory       map[string]([]*Pos) // history of at most 5 ticks
	Log               []byte
}

const (
	BOARD_SIZE       int    = 10
	CHECKIN_INTERVAL int    = 200
	DIRECTION_UP     string = "U"
	DIRECTION_DOWN   string = "D"
	DIRECTION_LEFT   string = "L"
	DIRECTION_RIGHT  string = "R"
	MAX_PLAYERS      int    = 6
	AXIS_X           int    = 0
	AXIS_Y           int    = 1
)

// Game variables.
var isPlaying bool        // Is the game in session.
var imAlive bool          // Am I alive.
var nodeId string         // Name of client.
var nodeIndex string      // Player number (1 - 6).
var nodeAddr string       // IP of client.
var httpServerAddr string // HTTP Server IP.
var nodes []*Node         // All nodes in the game.
var myNode *Node          // My node.

var HistoryLimit int              // Size limit for both nodeHistory and gameHistory
var nodeHistory map[string][]*Pos // Id to list of 5 recent local locations of each player
var aliveNodes int                // Number of alive nodes.

// #LEADER specific.
var failedNodes []string          // id of failed nodes found.
var gameHistory map[string][]*Pos // Last five moves of every node in the game. Written ONLY by the leader.

// Sync variables.
var waitGroup sync.WaitGroup // For internal processes.
var mutex *sync.Mutex        // For global vars.

// Game timers in milliseconds.
var intervalUpdateRate time.Duration
var tickRate time.Duration
var enforceGameStateRate time.Duration

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
	checkErr(err, 96)
	httpServerAddr = httpServerTcpAddr.String()

	log.Println(nodeAddr, nodeRpcAddr, msServerAddr, httpServerAddr)
	initLogging()

	waitGroup.Add(2) // Add internal process.
	go httpServe()
	go msRpcServe()
	waitGroup.Wait() // Wait until processes are done.
}

// Initialize variables.
func init() {
	directions = map[string]string{
		"p1": DIRECTION_RIGHT,
		"p2": DIRECTION_LEFT,
		"p3": DIRECTION_RIGHT,
		"p4": DIRECTION_LEFT,
		"p5": DIRECTION_RIGHT,
		"p6": DIRECTION_LEFT,
	}

	initialPosition = map[string]*Pos{
		"p1": &Pos{1, 1},
		"p2": &Pos{8, 8},
		"p3": &Pos{1, 8},
		"p4": &Pos{8, 1},
		"p5": &Pos{1, 4},
		"p6": &Pos{8, 5},
	}

	for player, pos := range initialPosition {
		board[pos.Y][pos.X] = player
	}

	nodeHistory = make(map[string][]*Pos)
	nodes = make([]*Node, 0)

	mutex = &sync.Mutex{}

	HistoryLimit = 5
	gameHistory = make(map[string][]*Pos)
	lastCheckin = make(map[string]time.Time)
	failedNodes = make([]string, 0)
	tickRate = 500 * time.Millisecond
	intervalUpdateRate = 1000 * time.Millisecond // TODO we said it's 100 in proposal?
	enforceGameStateRate = 2000 * time.Millisecond
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
	// Find myself and init variables.
	for _, node := range nodes {
		node.CurrLoc = initialPosition[node.Id]
		node.Direction = directions[node.Id]
		node.IsAlive = true
		lastCheckin[node.Id] = time.Now()
	}

	// Remove the node IDs of non-present players from the board.
	for i := len(nodes) + 1; i <= MAX_PLAYERS; i++ {
		pos, ok := initialPosition[fmt.Sprintf("p%d", i)]
		if !ok {
			continue
		}

		board[pos.Y][pos.X] = ""
	}

	localLog("nodeId:", nodeId)
	localLog("----INITIAL STATE----")
	printBoard()
	localLog("----INITIAL STATE----")

	// ================================================= //

	imAlive = true
	isPlaying = true
	aliveNodes = len(nodes)

	go listenUDPPacket()
	go intervalUpdate()
	go tickGame()
	go handleNodeFailure()
	go enforceGameState()
}

// Update the board based on leader's history
func UpdateBoard() {
	mutex.Lock()
	fmt.Println("Updating Board")
	localLog("Received gameHistory from Leader")

	// Clear everything on the board except our head
	for _, v := range nodeHistory {
		for _, e := range v {
			board[e.Y][e.X] = ""
		}
	}
	// Color board based on Leader's hitory
	for id, _ := range gameHistory {
		buf := []byte(id)
		playerIndex := string(buf[1])

		localLog(id)
		// for _, p := range nodeHistory[id] {
		for _, p := range gameHistory[id] {
			localLog(*p)
		}

		// Apply Leader's History onto the board
		for i, pos := range gameHistory[id] {
			if i == 0 {
				// Check if History's head is the same as our head
				board[pos.Y][pos.X] = getPlayerState(id)
				peerNode := getNode(id)
				peerNode.CurrLoc.X = pos.X
				peerNode.CurrLoc.Y = pos.Y
			} else {
				board[pos.Y][pos.X] = "t" + playerIndex
			}
		}
	}
	mutex.Unlock()
}

// Each tick of the game
func tickGame() {
	for {
		if isPlaying {
			mutex.Lock()
			for _, node := range nodes {
				playerIndex := string(node.Id[len(node.Id)-1])
				direction := node.Direction
				x := node.CurrLoc.X
				y := node.CurrLoc.Y
				new_x := node.CurrLoc.X
				new_y := node.CurrLoc.Y

				// only predict for live nodes
				if isPlaying && node.IsAlive {
					// Path prediction
					board[y][x] = "t" + playerIndex // Change position to be a trail.
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
						localLog("NODE " + node.Id + " IS DEAD")
						if isLeader() && node.Id == nodeId && node.IsAlive {
							node.IsAlive = false
							aliveNodes = aliveNodes - 1
							localLog("IM LEADER AND IM DEAD REPORTING TO FRONT END")
							notifyPlayerDeathToJS()
							reportASorrowfulDeathToPeers(node)
						} else if isLeader() {
							// we tell peers who the dead node is.
							node.IsAlive = false
							aliveNodes = aliveNodes - 1
							localLog("Leader sending death report ", node.Id)
							reportASorrowfulDeathToPeers(node)
						}
						// We don't update the position to a new value
						board[y][x] = getPlayerState(node.Id)
						if haveIWon() {
							localLog("Leader won")
							break
						}
					} else {
						// Update player's new position.
						board[new_y][new_x] = getPlayerState(node.Id)
						node.CurrLoc.X = new_x
						node.CurrLoc.Y = new_y
					}
				}
			}
			mutex.Unlock()
		}
		renderGame()
		time.Sleep(tickRate)
	}
}

// Change Position of a node by creating a trail from its previous location.
// (Predicting a path from a given prev location and new location).
func updateLocationOfNode(fromCurrent *Node, to *Node) {
	currentDir := fromCurrent.Direction
	newDir := to.Direction

	if newDir == currentDir {
		return
	}

	if currentDir == DIRECTION_UP || currentDir == DIRECTION_DOWN {
		matchPositionInAxis(AXIS_Y, false, fromCurrent, to)
		matchPositionInAxis(AXIS_X, true, fromCurrent, to)
	} else {
		matchPositionInAxis(AXIS_X, false, fromCurrent, to)
		matchPositionInAxis(AXIS_Y, true, fromCurrent, to)
	}

	fromCurrent.Direction = newDir
}

// Match position of current node to the new position in the
// given axis direction and whether to draw or delete trail.
// Axis is one of:
//		- AXIS_X
//		- AXIS_Y
func matchPositionInAxis(axis int, draw bool, from *Node, to *Node) {
	nodeName := from.Id
	fromX := from.CurrLoc.X
	fromY := from.CurrLoc.Y
	toX := to.CurrLoc.X
	toY := to.CurrLoc.Y

	nodeTrail := ""
	if draw {
		nodeTrail = "t" + nodeName[len(nodeName)-1:]
	}

	nodePlayer := getPlayerState(from.Id)

	if axis == AXIS_X { // Match X axis.
		i := fromX
		increment := -1
		if toX > fromX {
			increment = 1
		}
		for i != toX {
			board[fromY][i] = nodeTrail
			i = increment + i
		}
		board[fromY][i] = nodePlayer
		from.CurrLoc.X = toX
	} else { // Match Y axis.
		i := fromY
		increment := -1
		if toY > fromY {
			increment = 1
		}
		for i != toY {
			board[fromY][i] = nodeTrail
			i = increment + i
		}
		board[i][fromX] = nodePlayer
		from.CurrLoc.Y = toY
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
	mutex.Lock()
	if isLeader() {
		go collectLast7Moves()
	} else {
		// Only non-leader nodes have to do this
		go cacheLocation()
	}
	printBoard()
	pushGameStateToJS(board)
	mutex.Unlock()
}

// NON-LEADER: Build a history of last 5 moves for node on the board.
func cacheLocation() {
	mutex.Lock()
	// Collect the state of nodes on the board as the 'TRUE' state.
	for _, node := range nodes {
		// Clear the list.
		nodeHistory[node.Id] = make([]*Pos, 0)

		// Put in current location.
		nodeHistory[node.Id] = append(nodeHistory[node.Id], node.CurrLoc)

		i := 1
		xPos := node.CurrLoc.X
		yPos := node.CurrLoc.Y
		trail := "t" + string(node.Id[len(node.Id)-1])
		for i < 5 {
			p := findTrail(xPos, yPos, trail, nodeHistory[node.Id])
			if p != nil {
				nodeHistory[node.Id] = append(nodeHistory[node.Id], p)
				xPos = p.X
				yPos = p.Y
			} else {
				break
			}
			i++
		}

		localLog("Cache of node ", node.Id, "with len", len(nodeHistory[node.Id]))
		for _, p := range nodeHistory[node.Id] {
			localLog(*p)
		}
	}
	mutex.Unlock()
}

// LEADER: Build a history of last 7 moves for node on the board.
func collectLast7Moves() {
	// Collect the state of nodes on the board as the 'TRUE' state.
	for _, node := range nodes {
		// Clear the list.
		gameHistory[node.Id] = make([]*Pos, 0)

		// Put in current location.
		gameHistory[node.Id] = append(gameHistory[node.Id], node.CurrLoc)

		i := 1
		xPos := node.CurrLoc.X
		yPos := node.CurrLoc.Y
		trail := "t" + string(node.Id[len(node.Id)-1])
		for i < 7 {
			p := findTrail(xPos, yPos, trail, gameHistory[node.Id])
			if p != nil {
				gameHistory[node.Id] = append(gameHistory[node.Id], p)
				xPos = p.X
				yPos = p.Y
			} else {
				break
			}
			i++
		}

		localLog("History of node ", node.Id, "with len", len(gameHistory[node.Id]))
		for _, p := range gameHistory[node.Id] {
			localLog(*p)
		}
	}
}

// Find the next unvisited trail around the x, y position on the board.
// Return nil if trail cannot be found.
func findTrail(x int, y int, trail string, visited []*Pos) *Pos {
	if y > 0 && board[y-1][x] == trail && !contains(x, y-1, visited) {
		return &Pos{X: x, Y: y - 1}
	} else if y < BOARD_SIZE && board[y+1][x] == trail && !contains(x, y+1, visited) {
		return &Pos{X: x, Y: y + 1}
	} else if x > 0 && board[y][x-1] == trail && !contains(x-1, y, visited) {
		return &Pos{X: x - 1, Y: y}
	} else if x < BOARD_SIZE && board[y][x+1] == trail && !contains(x+1, y, visited) {
		return &Pos{X: x + 1, Y: y}
	} else {
		return nil
	}
}

// Check if x y is a position already in the list.
func contains(x int, y int, list []*Pos) bool {
	for _, p := range list {
		if p.X == x && p.Y == y {
			return true
		}
	}
	return false
}

// Continuously send game history of at most 5 previous ticks to all nodes
// Do it even if game ends because the last standing node might not communicate to other peers
func enforceGameState() {
	for {
		time.Sleep(enforceGameStateRate)
		if isLeader() {
			mutex.Lock()
			message := &Message{IsLeader: true, GameHistory: gameHistory, Node: *myNode}
			mutex.Unlock()
			logMsg := "Leader enforcing game state packet with game history"
			sendPacketsToPeers(logMsg, message)
			localLog(logMsg, message)
		}
	}
}

// Update peers with node's current location.
func intervalUpdate() {
	for {
		if imAlive == false || isPlaying == false {
			return
		}
		var message *Message
		if isLeader() {
			mutex.Lock()
			message = &Message{IsLeader: true, FailedNodes: failedNodes, Node: *myNode}
			mutex.Unlock()
		} else {
			message = &Message{Node: *myNode}
		}
		logMsg := "Interval update"
		sendPacketsToPeers(logMsg, message)
		time.Sleep(intervalUpdateRate)
	}
}

func sendPacketsToPeers(logMsg string, message *Message) {
	for _, node := range nodes {
		if node.Id != nodeId {
			log := logSend("Sending: " + logMsg + " [to: " + node.Id + " at ip " + node.Ip + "]")
			message.Log = log
			nodeJson, err := json.Marshal(message)
			checkErr(err, 548)
			go sendUDPPacket(node.Ip, nodeJson)
		}
	}
}

// Send data to ip via UDP.
func sendUDPPacket(ip string, data []byte) {
	// a random port is picked since we can't listen and read at the same time
	udpConn, err := net.Dial("udp", ip)
	checkErr(err, 559)
	defer udpConn.Close()

	_, err = udpConn.Write(data)
	checkErr(err, 563)
}

func processPacket(buf []byte, addr *net.UDPAddr, n int) {
	var message Message
	var node Node
	err := json.Unmarshal(buf[0:n], &message)
	checkErr(err, 570)
	node = message.Node

	logReceive("Received packet from "+addr.String()+": "+string(buf[0:n]), message.Log)
	localLog("Received: Id:", node.Id, "Ip:", node.Ip, "X:",
		node.CurrLoc.X, "Y:", node.CurrLoc.Y, "Dir:", node.Direction)
	lastCheckin[node.Id] = time.Now()

	if message.IsLeader {
		// FailedNodes communication.
		if message.FailedNodes != nil {
			localLog("failedNodes are: ", message.FailedNodes)
			for _, n := range message.FailedNodes {
				removeNodeFromList(n)
			}
		}

		// Check if message.History exist
		if message.GameHistory != nil {
			// Cache history info from the leader
			gameHistory = message.GameHistory
			UpdateBoard()
		}
	}

	if message.IsDeathReport {
		localLog("Received death report ", node.Id)
		mutex.Lock()
		// update local copy
		for _, n := range nodes {
			if n.Id == node.Id && n.IsAlive {
				n.IsAlive = false
				localLog("LEADER SENT: ", n.Id, " IS DEAD")
				aliveNodes = aliveNodes - 1
				localLog("**** DEATH REPORT *** size is now ", strconv.Itoa(aliveNodes))
				board[n.CurrLoc.Y][n.CurrLoc.X] = getPlayerState(n.Id)

				// Check if its me.
				if node.Id == nodeId {
					localLog("OH SHOOT ITS ME")
					notifyPlayerDeathToJS()
				}
			}
		}

		if haveIWon() {
			mutex.Unlock()
			renderGame()
			return
		}
		mutex.Unlock()
	}

	// Received a direction change from a peer.
	// Match the state of peer by predicting its path.
	if message.IsDirectionChange {
		mutex.Lock()
		for _, n := range nodes {
			if n.Id == message.Node.Id {
				n.Direction = message.Node.Direction
			}
		}
		mutex.Unlock()
	}
	mutex.Lock()
	mNode := getNode(message.Node.Id)
	updateLocationOfNode(mNode, &message.Node)
	mutex.Unlock()
}

func listenUDPPacket() {
	localAddr, err := net.ResolveUDPAddr("udp", nodeAddr)
	checkErr(err, 642)
	udpConn, err := net.ListenUDP("udp", localAddr)
	checkErr(err, 644)
	err = udpConn.SetReadBuffer(9000)
	checkErr(err, 646)
	defer udpConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := udpConn.ReadFromUDP(buf)
		checkErr(err, 653)
		go processPacket(buf, addr, n)
		time.Sleep(100 * time.Millisecond)
	}
}

// LEADER: Tell nodes someone has died.
func reportASorrowfulDeathToPeers(node *Node) {
	msg := &Message{IsDeathReport: true, Node: *node}
	logMsg := "Node " + node.Id + "is dead, reporting sorrowful death"
	sendPacketsToPeers(logMsg, msg)
}

func haveIWon() bool {
	// stop playing when someone has won
	if myNode.IsAlive && aliveNodes == 1 {
		localLog("I WIN")
		notifyPlayerVictoryToJS()
		isPlaying = false
		return true
	} else if aliveNodes == 1 {
		localLog("Someone else won")
		isPlaying = false
	}
	return false
}

func notifyPeersDirChanged(direction string) {
	mutex.Lock()
	prevDirection := myNode.Direction

	// check if the direction change for node with the id
	if prevDirection != direction {
		logMsg := "Direction for " + nodeId + " has changed from " +
			prevDirection + " to " + direction
		myNode.Direction = direction

		msg := &Message{IsDirectionChange: true, Node: *myNode}
		localLog(logMsg, msg)
		sendPacketsToPeers(logMsg, msg)
	}
	mutex.Unlock()
}

func isLeader() bool {
	return nodes[0].Id == nodeId
}

func hasExceededThreshold(nodeLastCheckin int64) bool {
	// TODO gotta check the math : fix incoming.
	threshold := nodeLastCheckin + (7000 * int64(time.Millisecond/time.Nanosecond))
	now := time.Now().UnixNano()
	return threshold < now
}

func handleNodeFailure() {
	// check if the time it last checked in exceed CHECKIN_INTERVAL
	for {
		if isPlaying == false {
			return
		}
		if isLeader() {
			localLog("Im a leader: ", nodeId)
			for _, node := range nodes {
				if node.Id != nodeId {
					if hasExceededThreshold(lastCheckin[node.Id].UnixNano()) {
						localLog(node.Id, " HAS FAILED")
						// --> leader should periodically send out active nodes in the system
						// --> so here we just have to remove it from the nodes list.
						failedNodes = append(failedNodes, node.Id)
						localLog(len(failedNodes))
						removeNodeFromList(node.Id)
					}
				}
			}
		} else {
			localLog("Im a node: ", nodeId)
			// Continually check if leader is alive.
			leaderId := nodes[0].Id
			if hasExceededThreshold(lastCheckin[leaderId].UnixNano()) {
				localLog("LEADER ", leaderId, " HAS FAILED.")
				removeNodeFromList(leaderId)
			}
		}
		time.Sleep(intervalUpdateRate)
	}
}

// LEADER: removes a dead node from the node list.
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

// Given a node id string, return "p_" or "d_" depending on whether the player is alive.
func getPlayerState(id string) string {
	for _, n := range nodes {
		if n.Id == id {
			buf := []byte(id)
			playerIndex := string(buf[1])
			if n.IsAlive {
				return "p" + playerIndex
			} else {
				return "d" + playerIndex
			}
		}
	}
	return ""
}

// Given a node id string, return the node's location X, Y on the board.
func getPlayerLocation(id string) (int, int) {
	for _, n := range nodes {
		if n.Id == id {
			return n.CurrLoc.X, n.CurrLoc.Y
		}
	}
	return 0, 0
}

// Given a node id string, return the node.
func getNode(id string) *Node {
	for _, n := range nodes {
		if n.Id == id {
			return n
		}
	}
	return nil
}

// Error checking. Exit program when error occurs.
func checkErr(err error, lineNum int) {
	if err != nil {
		localLog("line ", strconv.Itoa(lineNum)+", error:", err)
		os.Exit(1)
	}
}

// For debugging
func printBoard() {
	// TODO: Continous string concat is terrible, but this is OK for just
	//       debugging for now. Get rid of it at some point in the future.
	topLine := "  "
	for i, _ := range board[0] {
		topLine += fmt.Sprintf("%3d", i)
	}
	localLog(topLine, "")
	for r, _ := range board {
		// Ideally, we would introduce a localLog() variant that does Print()
		// instead of Println(). However, Print() on log files seems to always
		// introduce a new line, which is useless for what we're doing here.
		line := ""
		for _, item := range board[r] {
			if item == "" {
				line += "__ "
			} else {
				line += (item + " ")
			}
		}
		localLog(fmt.Sprintf("%2d", r), line)
	}
}
