package main

import (
	"log"
	"os"
)

type GameState struct {
	board Board
	// where the players are now
	currLocs map[string]Location
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
	dir int
}

const (
	boardSize int = 10
)

var gameState GameState
var done chan int

func main() {
	if len(os.Args) < 5 {
		log.Println("usage: node.go [nodeid] [nodeAddr] [msServerAddr] [httpServerAddr]")
		log.Println("[nodeid] unique id for each player")
		log.Println("[nodeAddr] the rpc ip:port the node is listening to")
		log.Println("[msServerAddr] the rpc ip:port of matchmaking server is listening to")
		log.Println("[httpServerAddr] the ip:port the http server is binded to ")
		os.Exit(1)
	}

	nodeId, nodeAddr, msServerAddr, httpServerAddr := os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	log.Println(nodeId, nodeAddr, msServerAddr, httpServerAddr)

	initLogging()

	httpMsg := make(chan string)
	rpcMsg := make(chan string)
	go msRpcServce(rpcMsg, msServerAddr)
	go httpServe(httpMsg)
	<-done
	<-done
}

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
	log.Println(gameState)
}
