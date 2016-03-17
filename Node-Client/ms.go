package main

import (
	"log"
	"net"
	"net/rpc"
)

type NodeClient int

type ValReply struct {
	msg string
}

type GameArgs struct {
	nodeList []Node
}

var nodeRpcAddr string
var msServerAddr string // Matchmaking server IP.
var msService *rpc.Client

// This RPC function is triggered when a game is ready to begin.
func (nc *NodeClient) StartGame(args *GameArgs, response *ValReply) error {
	nodes = args.nodeList
	log.Println("Starting game with nodes:")
	// startGame() // in node.go, call when rpc is working
	return nil
}

// This RPC function serves as a way for the Matchmaking service to send text to this node.
func (nc *NodeClient) Message(args *GameArgs, response *ValReply) error {
	log.Println("Received message:" + response.msg)
	return nil
}

func msRpcServce() {
	defer waitGroup.Done()
	nodeClient := new(NodeClient)
	rpc.Register(nodeClient)
	nodeListener, e := net.Listen("tcp", nodeRpcAddr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	log.Println("Listening for ms server at ", nodeRpcAddr)
	conn, _ := nodeListener.Accept()
	rpc.ServeConn(conn)
}

func connectMs() {
	msService, err := rpc.Dial("tcp", msServerAddr)
	if err != nil {
		log.Fatal("connect error:", err)
	}
	log.Println("Connected to matchmaking server", msService)
}
