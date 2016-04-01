package main

import (
	"github.com/arcaneiceman/GoVector/govec"
	"log"
	"os"
	"strings"
	"time"
)

type Msg struct {
	RealTimestamp string
}

var Logger *govec.GoLog
var fileLogger *log.Logger

func initLogging(rpcAddr string) {
	// Windows doesn't accept colons in paths, so we filter them out here.
	logFileName := strings.Replace(rpcAddr, ":", "", -1)
	Logger = govec.Initialize(rpcAddr, logFileName)

	localLogFile, err := os.Create(logFileName + "-local.txt")
	CheckError(err, 24)
	fileLogger = log.New(localLogFile, "", 0)
}

func logSend(msg string) []byte {
	outgoingMessage := Msg{time.Now().String()}
	return Logger.PrepareSend(msg, outgoingMessage)
}

func logReceive(msg string, buf []byte) *Msg {
	incommingMessage := new(Msg)
	Logger.UnpackReceive(msg, buf[:], &incommingMessage)
	return incommingMessage
}

func localLog(v ...interface{}) {
	log.Println(v)
	fileLogger.Println(v)
}
