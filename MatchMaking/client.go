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

type MatchMakingService int

// Reply from MS server
type ValReply struct {
	Val string // value; depends on the call
}

type GameArgs struct {
	nodeList []string
}

func (kvs *MatchMakingService) StartGame(args *GameArgs, reply *ValReply) error {
	log.Println("Starting game")
	return nil
}

type Node struct {
	Id string // Napon
	Ip string // ip addr of Napon
}

func main() {
	// go run client.go clientIP msIPport
	// go run client.go :4421 :4431
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
		// Export MatchMakingService to allow MS to trigger start game here
		msService := new(MatchMakingService)
		rpc.Register(msService)
		msListener, e := net.Listen("tcp", localAddr.String())
		if e != nil {
			log.Fatal("listen error:", e)
		}
		log.Println("Listening to ms server")
		conn, _ := msListener.Accept()
		rpc.ServeConn(conn)
	}()

	// First connect to the MS server
	client, e := rpc.Dial("tcp", remoteAddr.String())
	FatalError(e)
	defer client.Close()

	var reply *ValReply = &ValReply{Val: ""}
	e = client.Call(RpcJoin, &Node{Id: os.Args[3], Ip: localAddr.String()}, reply)
	CheckError(e, 6)

	fmt.Println("Reply: ", reply.Val)
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
