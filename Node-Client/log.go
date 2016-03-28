package main

import (
	"github.com/arcaneiceman/GoVector/govec"
	"time"
)

type Msg struct {
	Payload       []byte
	RealTimestamp string
}

var Logger *govec.GoLog

func initLogging() {
	Logger = govec.Initialize(nodeAddr, nodeAddr)
}

func send(msg string, payload []byte) []byte {
	outgoingMessage := Msg{payload, time.Now().String()}
	return Logger.PrepareSend(msg, outgoingMessage)
}

func receive(msg string, buf []byte, n int) *Msg {
	incommingMessage := new(Msg)
	Logger.UnpackReceive(msg, buf[0:n], &incommingMessage)
	return incommingMessage
}
