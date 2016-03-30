package main

import (
	"github.com/arcaneiceman/GoVector/govec"
	"log"
	"os"
	"strings"
	"time"
)

type Msg struct {
	Payload       []byte
	RealTimestamp string
}

var Logger *govec.GoLog
var fileLogger *log.Logger

func initLogging() {
	// Windows doesn't accept colons in paths, so we filter them out here.
	logFileName := strings.Replace(nodeAddr, ":", "", -1)
	Logger = govec.Initialize(nodeAddr, logFileName)

	localLogFile, err := os.Create(logFileName + "-local.txt")
	checkErr(err)
	fileLogger = log.New(localLogFile, "", 0)
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

func localLog(v ...interface{}) {
	log.Println(v)
	fileLogger.Println(v)
}
