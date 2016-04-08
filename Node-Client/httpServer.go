package main

import (
	"github.com/googollee/go-socket.io"
	"github.com/pkg/browser"
	"log"
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
	log.Println("Start game ", nodeId+" "+nodeAddr)
	_gSO.Emit("startGame", nodeId+" "+nodeAddr)
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

	_gSO.Emit("victory")
}

func httpServe() {
	defer waitGroup.Done()
	server, err := socketio.NewServer(nil)
	if err != nil {
		localLog("ERROR: Fatal socketio.NewServer() error:", err)
		log.Fatal(err)
	}
	// TODO: For some weird reason, "connection" is invoked many times when there are multiple browser windows all pointing to the same localhost UI port, causing the UI to not start properly. A dumb fix for this is to only allow ONE connection.
	server.SetMaxConnection(1)
	server.On("connection", func(so socketio.Socket) {
		localLog("on connection")
		_gSO = so
	})
	server.On("error", func(so socketio.Socket, err error) {
		localLog("ERROR:", err)
	})
	server.On("disconnection", func() {
		localLog("on disconnect")
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at ", httpServerAddr, "...")

	// Point the default browser at the page to save the user the effort of
	// doing it themselves.
	// TODO: Currently, as a hack, we tell the default browser to open up the
	//       URL for where the server will be listening at, and hope the server
	//       wins the race. We should really be doing this only after we know
	//       the server is fully up.
	browser.OpenURL("http://" + httpServerAddr)

	log.Fatal(http.ListenAndServe(httpServerAddr, nil))
}
