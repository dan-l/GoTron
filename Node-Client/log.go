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

func initLogging() {
	// Windows doesn't accept colons in paths, so we filter them out here.
	logFileName := strings.Replace(nodeAddr, ":", "", -1)
	Logger = govec.Initialize(nodeAddr, logFileName)

	localLogFile, err := os.Create(logFileName + "-local.txt")
	checkErr(err)
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
	fileLogger.Println(time.Now().Format("2006-01-02 15:04:05.000"), v)
}
