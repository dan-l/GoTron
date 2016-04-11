package main

// This file implements communication with a matchmaking server on the client
// side.

import (
	"errors"
	"net"
	"net/rpc"
	"strconv"
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
	logReceive("Rpc Called Start Game to "+msServerAddr, args.Log)
	if len(nodes) > MAX_PLAYERS {
		return errors.New("MS Server returned a node list with more than the " +
			"max number of supported players")
	}

	localLog("Starting game with nodes: " + printNodes())
	findMyNode()
	msService.Close()
	startGame()   // in node.go, call when rpc is working
	startGameUI() // in httpServer.go, transition to game screen on the client.
	return nil
}

func findMyNode() {
	for i, node := range nodes {
		if node.Ip == nodeAddr {
			myNode = node
			nodeId = node.Id
			nodeIndex = strconv.Itoa(i + 1)
		}
	}
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
	localLog("Received message:" + response.Val)
	return nil
}

func msRpcServe() {
	defer waitGroup.Done()

	localAddr, err := net.ResolveTCPAddr("tcp", nodeRpcAddr)
	checkErr(err, 78)

	nodeService := new(NodeService)
	rpc.Register(nodeService)
	nodeListener, err := net.Listen("tcp", localAddr.String())
	checkErr(err, 83)

	localLog("Listening for ms server at ", localAddr.String())
	conn, err := nodeListener.Accept()
	checkErr(err, 87)
	go rpc.ServeConn(conn)
}

func msRpcDial() {
	remoteAddr, e := net.ResolveTCPAddr("tcp", msServerAddr)
	checkErr(e, 93)

	msService, e = rpc.Dial("tcp", remoteAddr.String())
	checkErr(e, 96)

	var reply *ValReply = &ValReply{Val: ""}
	log := logSend("Rpc Call Context.Join to " + msServerAddr)
	err := msService.Call("Context.Join",
		&NodeJoin{RpcIp: nodeRpcAddr, Ip: nodeAddr, Log: log}, reply)
	checkErr(err, 101)
}
