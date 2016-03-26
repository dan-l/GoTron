package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	//"sync"
	//"time"
)

const RpcJoin string = "Context.Join"

func CheckError(err error, n int) {
	if err != nil {
		fmt.Println(n, ": ", err)
		os.Exit(n)
	}
}

type NodeService int

func (kvs *NodeService) StartGame(args *GameArgs, reply *ValReply) error {
	log.Println("Starting game")
	fmt.Println("nodelist:", args.NodeList)
	reply.Val = "cool"
	return nil
}

// This RPC function serves as a way for the Matchmaking service to send text to this node.
func (nc *NodeService) Message(args *GameArgs, response *ValReply) error {
	log.Println("Received message:", args.NodeList)
	response.Val = "ok"
	return nil
}

// Reply from MS server
type ValReply struct {
	Val string // value; depends on the call
}

type GameArgs struct {
	NodeList []*Node
}

type Node struct {
	Id string // Napon
	Ip string // ip addr of Napon
}

type NodeJoin struct {
	RpcIp string // The one MS has to dial at start Game
	Ip    string // ip to send to each player
}

func main() {
	// go run client.go rpcIP msIPport nodeIP
	// go run client.go :4101 :4421 :4441
	if len(os.Args) != 4 {
		fmt.Println("Not enough arguments")
		os.Exit(-1)
	}

	// Setting arguments
	localAddr, e := net.ResolveTCPAddr("tcp", os.Args[1])
	FatalError(e)

	remoteAddr, e := net.ResolveTCPAddr("tcp", os.Args[2])
	FatalError(e)

	//Exporting methods to be used by MS server
	go func() {
		nodeService := new(NodeService)
		rpc.Register(nodeService)
		nodeListener, e := net.Listen("tcp", localAddr.String())
		if e != nil {
			log.Fatal("listen error:", e)
		}
		log.Println("Listening for ms server at ", localAddr.String())
		conn, _ := nodeListener.Accept()
		rpc.ServeConn(conn)
	}()

	// First connect to the MS server
	client, e := rpc.Dial("tcp", remoteAddr.String())
	FatalError(e)
	defer client.Close()

	var reply *ValReply = &ValReply{Val: ""}
	e = client.Call(RpcJoin, &NodeJoin{RpcIp: localAddr.String(), Ip: os.Args[3]}, reply)
	CheckError(e, 6)

	fmt.Println("Reply: ", reply.Val)
	for {

	}
}

// Level for printing
// 0 - only errors
// 1 - general connection info, key info
// 2 - message aggreagtion
// 3 - Messages being sent
// 4 - Everything
const DebugLevel int = 4

func DebugPrint(level int, str string) {
	if level <= DebugLevel {
		fmt.Println(str)
	}
}

// check if a fatal error has ocurred
func FatalError(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(-10)
	}
}
