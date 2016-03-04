package main

import (
	"github.com/arcaneiceman/GoVector/govec"
	"time"
)

type Msg struct {
	Content, RealTimestamp string
}

var Logger *govec.GoLog

func initLogging() {
	Logger = govec.Initialize("GoTron", "GoTron")
}

func send(msg string) []byte {
	outgoingMessage := Msg{msg, time.Now().String()}
	return Logger.PrepareSend("Sending message to server", outgoingMessage)
}

func receive(msg string) *Msg {
	incommingMessage := new(Msg)
	var buf [512]byte
	Logger.UnpackReceive("Received Message From Client", buf[0:], &incommingMessage)
	return incommingMessage
}
