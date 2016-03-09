package main

import (
	"log"
	"net"
	"net/rpc"
)

type NodeService int

type ValReply struct {
	msg string
}

type GameArgs struct {
	nodeList []string
}

var nodeRpcAddr string
var msServerAddr string // Matchmaking server IP.
var msService *rpc.Client

func (kvs *NodeService) StartGame(args *GameArgs, reply *ValReply) error {
	log.Println("Starting game")
	return nil
}

func msRpcServce() {
	defer waitGroup.Done()
	nodeService := new(NodeService)
	rpc.Register(nodeService)
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

func notifyPeersDirChanged() {
	// msService.Call("DirectionChanged")
	// listen to socket io for direction changed topic
}
