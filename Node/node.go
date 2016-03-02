package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("usage: node.go [nodeAddr] [msServerAddr] [httpServerAddr]")
		os.Exit(1)
	}

	nodeAddr, msServerAddr, httpServerAddr := os.Args[1], os.Args[2], os.Args[3]

	fmt.Println(nodeAddr, msServerAddr, httpServerAddr)
}
