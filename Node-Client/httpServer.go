package main

import (
	"github.com/googollee/go-socket.io"
	"github.com/pkg/browser"
	"log"
	"net/http"
)

// A fake board for testing/bring up purposes.
// TODO: Remove this once the necessary bring up is done.
// TODO: The real board will use constants such as "p1" to track current player
//       locations. We should handle that.
var httpServerFakeBoard [BOARD_SIZE][BOARD_SIZE]string = [BOARD_SIZE][BOARD_SIZE]string{
	[BOARD_SIZE]string{"TEMP USER ID", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
}
var playerPos Pos = Pos{0, 0}

func intMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func intMin(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// Updates the internal state of the board, then returns the new state.
func updateInternalState(direction string) [BOARD_SIZE][BOARD_SIZE]string {
	var playerID string = httpServerFakeBoard[playerPos.Y][playerPos.X]

	// Clear the current position.
	// TODO: Implement trails.
	httpServerFakeBoard[playerPos.Y][playerPos.X] = ""

	if direction == "U" {
		playerPos.Y = intMax(0, playerPos.Y-1)
	} else if direction == "D" {
		playerPos.Y = intMin(BOARD_SIZE-1, playerPos.Y+1)
	} else if direction == "L" {
		playerPos.X = intMax(0, playerPos.X-1)
	} else if direction == "R" {
		playerPos.X = intMin(BOARD_SIZE-1, playerPos.X+1)
	}

	// Set the new position.
	httpServerFakeBoard[playerPos.Y][playerPos.X] = playerID

	return httpServerFakeBoard
}

func httpServe() {
	defer waitGroup.Done()
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
		so.Join("chat")
		so.On("playerMove", func(playerMove map[string]string) {
			_, ok := playerMove["id"]
			if !ok {
				// TODO Output error message somewhere
				return
			}

			direction, ok := playerMove["direction"]
			if !ok {
				// TODO Output error message somewhere
				return
			}

			so.Emit("gameStateUpdate", updateInternalState(direction))
		})
		so.On("disconnection", func() {
			log.Println("on disconnect")
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
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
