package main

import (
	"github.com/googollee/go-socket.io"
	"github.com/pkg/browser"
	"log"
	"net/http"
)

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

			_, ok = playerMove["direction"]
            if !ok {
                // TODO Output error message somewhere
                return
            }

            // TODO: This exists as an interim step to getting full game state
            //       transmitted back to the JS layer, and should be removed
            //       later.
            so.Emit("playerMoveEcho", playerMove);
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
