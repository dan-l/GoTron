package main

import (
	"log"
	"net"
	"net/rpc"
)

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

func msRpcServce(messages chan string, msIpPort string) {
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
