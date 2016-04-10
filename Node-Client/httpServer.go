package main

import (
	"github.com/googollee/go-socket.io"
	"github.com/pkg/browser"
	"log"
	"net"
	"net/http"
)

var playerID string

// Note: This variable should be treated as private to httpServer.go.
var _gSO socketio.Socket

// Starts the UI game screen.
func startGameUI() {
	if _gSO == nil {
		localLog("socketio is NIL !!!")
		return
	}

	_gSO.On("playerMove", func(playerMove map[string]string) {
		direction, ok := playerMove["direction"]
		if !ok {
			log.Fatal("Received playerMove without direction")
			return
		}

		notifyPeersDirChanged(direction)
	})

	// Start the game.
	_gSO.Emit("startGame", nodeId, nodeAddr, myNode.Direction)
}

func pushGameStateToJS(state [BOARD_SIZE][BOARD_SIZE]string) {
	if _gSO == nil {
		localLog("socketio is NIL !!!")
		return
	}

	_gSO.Emit("gameStateUpdate", state)
}

func notifyPlayerDeathToJS() {
	if _gSO == nil {
		localLog("socketio is NIL !!!")
		return
	}

	_gSO.Emit("playerDead")
}

func notifyPlayerVictoryToJS() {
	if _gSO == nil {
		localLog("socketio is NIL !!!")
		return
	}

	_gSO.Emit("playerVictory")
}

func httpServe() {
	defer waitGroup.Done()
	server, err := socketio.NewServer(nil)
	if err != nil {
		localLog("ERROR: Fatal socketio.NewServer() error:", err)
		log.Fatal(err)
	}
	server.On("connection", func(so socketio.Socket) {
		localLog("on connection")
		_gSO = so
		go msRpcDial()
	})
	server.On("error", func(so socketio.Socket, err error) {
		localLog("ERROR:", err)
	})
	server.On("disconnection", func() {
		localLog("on disconnect")
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	localLog("Serving at ", httpServerAddr, "...")

	listener, err := net.Listen("tcp", httpServerAddr)
	if err != nil {
		localLog("httpserver listener error ", err)
	} else {
		localLog("httpserver listener success")
		browser.OpenURL("http://" + httpServerAddr)
		http.Serve(listener, nil)
	}
}
