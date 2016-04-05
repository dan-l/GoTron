package main

import (
	"errors"
	"log"
	"net"
	"net/rpc"
)

type NodeService int

type ValReply struct {
	Val string
}

type GameArgs struct {
	NodeList []*Node
	Log      []byte
}

type NodeJoin struct {
	RpcIp string
	Ip    string
	Log   []byte
}

var nodeRpcAddr string
var msServerAddr string // Matchmaking server IP.
var msService *rpc.Client

// This RPC function is triggered when a game is ready to begin.
func (nc *NodeService) StartGame(args *GameArgs, response *ValReply) error {
	nodes = args.NodeList
	logReceive("Rpc Called Start Game", args.Log)
	if len(nodes) > MAX_PLAYERS {
		return errors.New("MS Server returned a node list with more than the " +
			"max number of supported players")
	}

	log.Println("Starting game with nodes: " + printNodes())
	msService.Close()
	startGame()   // in node.go, call when rpc is working
	startGameUI() // in httpServer.go, transition to game screen on the client.
	return nil
}

// Print node in the list
func printNodes() string {
	result := ""
	for _, n := range nodes {
		result += n.Ip + " "
	}
	return result
}

// This RPC function serves as a way for the Matchmaking service to send text to this node.
func (nc *NodeService) Message(args *GameArgs, response *ValReply) error {
	logReceive("Rpc Called Message", args.Log)
	log.Println("Received message:" + response.Val)
	return nil
}

func msRpcServce() {
	defer waitGroup.Done()

	localAddr, e := net.ResolveTCPAddr("tcp", nodeRpcAddr)
	checkErr(e)

	remoteAddr, e := net.ResolveTCPAddr("tcp", msServerAddr)
	checkErr(e)

	go func() {
		nodeService := new(NodeService)
		rpc.Register(nodeService)
		nodeListener, e := net.Listen("tcp", localAddr.String())
		checkErr(e)

		log.Println("Listening for ms server at ", localAddr.String())
		conn, _ := nodeListener.Accept()
		rpc.ServeConn(conn)
	}()

	msService, e = rpc.Dial("tcp", remoteAddr.String())
	checkErr(e)

	var reply *ValReply = &ValReply{Val: ""}
	log := logSend("Rpc Call Context.Join")
	_ = msService.Call("Context.Join", &NodeJoin{RpcIp: nodeRpcAddr, Ip: nodeAddr, Log: log}, reply)
}
