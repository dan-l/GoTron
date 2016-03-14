package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	//"sync"
	//"time"
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

type HelloMessage struct {
	Id string
}

func main() {
	// go run client.go clientIP msIPport
	// go run client.go :4421 :4431
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <ip:port>\n", os.Args[0])
		fmt.Println("\t<ip:port> - The IP and port of the MatchMaking rpc server to connect to")
		os.Exit(-1)
	}

	// Setting arguments
	clientIP := os.Args[1]
	msIpPort := os.Args[2]

	fmt.Println(clientIP, " and ", msIpPort)

	localAddr, e := net.ResolveTCPAddr("tcp", clientIP)
	FatalError(e)

	remoteAddr, e := net.ResolveTCPAddr("tcp", msIpPort)
	FatalError(e)

	//Exporting methods to be used by MS server
	go func() {
		// Export MatchMakingService to allow MS to trigger start game here
		msService := new(MatchMakingService)
		rpc.Register(msService)
		msListener, e := net.Listen("tcp", remoteAddr.String())
		if e != nil {
			log.Fatal("listen error:", e)
		}
		log.Println("Listening to ms server")
		conn, _ := msListener.Accept()
		rpc.ServeConn(conn)
	}()

	// First connect to the MS server
	conn, e := net.DialTCP("tcp", localAddr, remoteAddr)
	if e != nil {
		// failed connection, sleep and try again in a bit
		//DebugPrint(1, "Reconnecting to leader in "+string(ReconnectInterval))
		//time.Sleep(ReconnectInterval)
		fmt.Println("Fail to connect")
	}

	hello := &HelloMessage{Id: clientIP}
	bHello, e := json.Marshal(hello)
	FatalError(e)
	//fmt.Println("Client->MS:", bHello[:])

	_, e = conn.Write(bHello) // Send hello message to frontend
	FatalError(e)

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
