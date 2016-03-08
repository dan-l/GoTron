package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func CheckError(err error, n int) {
	if err != nil {
		fmt.Println(n, ": ", err)
		os.Exit(n)
	}
}

type MatchMakingService int

type ValReply struct {
	msg string
}

type GameArgs struct {
	nodeList []string
}

func (kvs *MatchMakingService) StartGame(args *GameArgs, reply *ValReply) error {
	log.Println("Starting game")
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <ip:port>\n", os.Args[0])
		fmt.Println("\t<ip:port> - The IP and port of the MatchMaking rpc server to connect to")
		os.Exit(-1)
	}

	// Setting arguments
	clentIP = os.Args[1]
	msIpPort = os.Args[2]

	// Export MatchMakingService to allow MS to trigger start game here
	msService := new(MatchMakingService)
	rpc.Register(msService)
	msListener, e := net.Listen("tcp", msIpPort)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	log.Println("Listening to ms server")
	conn, _ := msListener.Accept()
	rpc.ServeConn(conn)
	done <- 1
}
