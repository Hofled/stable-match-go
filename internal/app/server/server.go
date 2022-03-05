/*
A socket.io server for communicating with the client side
*/
package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Hofled/stable-matching-go/internal/app/types"
	socketio "github.com/googollee/go-socket.io"
)

func UpdateState(server *socketio.Server, state interface{}) bool {
	return server.BroadcastToNamespace("/", "state", state)
}

func UpdatePeople(server *socketio.Server, men []*types.Man, women []*types.Woman) bool {
	people := struct {
		Men   []*types.Man
		Women []*types.Woman
	}{
		Men:   men,
		Women: women,
	}
	return server.BroadcastToNamespace("/", "update-people", people)
}

func SetupHandler(server *socketio.Server, namespace, eventName string, f interface{}) {
	server.OnEvent(namespace, eventName, f)
}

// starts serving the server and returns it
func Serve(port int) *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(c socketio.Conn) error {
		c.SetContext("")
		fmt.Printf("New connection with ID %v\n", c.ID())
		return nil
	})

	go server.Serve()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("../../internal/app/client")))
	fmt.Printf("Serving on http://localhost:%d...\n", port)
	// serve
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()

	return server
}
