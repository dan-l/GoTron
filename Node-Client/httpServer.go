package main

import (
	"errors"
	"fmt"
	"github.com/googollee/go-socket.io"
	"github.com/pkg/browser"
	"log"
	"net/http"
	"strings"
)

// A fake board for testing/bring up purposes.
// TODO: Remove this once the necessary bring up is done.
var httpServerFakeBoard [BOARD_SIZE][BOARD_SIZE]string = [BOARD_SIZE][BOARD_SIZE]string{
	[BOARD_SIZE]string{"p1", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"", "", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"t6", "p6", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"t5", "p5", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"t4", "p4", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"t3", "p3", "", "", "", "", "", "", "", ""},
	[BOARD_SIZE]string{"t2", "p2", "", "", "", "", "", "", "", ""},
}
var playerPos Pos = Pos{0, 0}
var userCodeToID map[string]string = make(map[string]string)

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
func updateInternalState(direction string) ([BOARD_SIZE][BOARD_SIZE]string, error) {
	var playerCode string = httpServerFakeBoard[playerPos.Y][playerPos.X]
	if len(playerCode) != 2 {
		err := errors.New(fmt.Sprintln("Player code", playerCode, "is invalid"))
		return httpServerFakeBoard, err
	}

	// Replace the current position with a trail.
	trailPlayerCode := strings.Replace(playerCode, "p", "t", 1)
	httpServerFakeBoard[playerPos.Y][playerPos.X] = trailPlayerCode

	if direction == "U" {
		playerPos.Y = intMax(0, playerPos.Y-1)
	} else if direction == "D" {
		playerPos.Y = intMin(BOARD_SIZE-1, playerPos.Y+1)
	} else if direction == "L" {
		playerPos.X = intMax(0, playerPos.X-1)
	} else if direction == "R" {
		playerPos.X = intMin(BOARD_SIZE-1, playerPos.X+1)
	} else {
		err := errors.New(fmt.Sprintln("Direction", direction, "is invalid"))
		return httpServerFakeBoard, err
	}

	// Set the new position.
	httpServerFakeBoard[playerPos.Y][playerPos.X] = playerCode

	return httpServerFakeBoard, nil
}

type InitialConfig struct {
	Players map[string]string
}

// Sets up the initial config using the given player ID from the JS layer.
// Returns the corresponding config once done.
func setupInitialConfig(playerID string) InitialConfig {
	userCodeToID["p1"] = playerID
	return InitialConfig{Players: userCodeToID}
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
		so.On("playerInfo", func(playerInfo map[string]string) {
			id, ok := playerInfo["id"]
			if !ok {
				// TODO Output error message somewhere
				return
			}

			so.Emit("initialConfig", setupInitialConfig(id))
			// After sending the config, we need to send an initial game state
			// update so the JS layer renders everything.
			so.Emit("gameStateUpdate", httpServerFakeBoard)
		})
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

			state, err := updateInternalState(direction)
			if err != nil {
				// TODO Output error message somewhere
				return
			}

			so.Emit("gameStateUpdate", state)
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
