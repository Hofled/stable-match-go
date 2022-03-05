/*

Performs stable matching between a group of `n` men and women.
Preference ranking is in ascending order, which means lower value has a higher preference.

*/
package main

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Hofled/stable-matching-go/internal/app/algorithm"
	"github.com/Hofled/stable-matching-go/internal/app/server"
	"github.com/Hofled/stable-matching-go/internal/app/types"
	socketio "github.com/googollee/go-socket.io"
)

const default_group_size = 5

var (
	Verbose      bool
	GroupSize    int
	UnmarriedMen *list.List
	Women        []*types.Woman
)

func init() {
	flag.BoolVar(&Verbose, "v", false, "Whether to print the generated groups info")
	flag.IntVar(&GroupSize, "n", default_group_size, "The group size for each gender")
}

func TrackTime() func() time.Duration {
	start := time.Now()
	// returns the duration since we started tracking time
	return func() time.Duration {
		return time.Since(start)
	}
}

func main() {
	flag.Parse()

	// server socket-io server
	s := server.Serve(8000)
	defer s.Close()

	// setup handlers
	server.SetupHandler(s, "/", "generate", func(c socketio.Conn, groupSize int) {
		UnmarriedMen, Women = algorithm.GenerateGroups(s, groupSize, Verbose)
	})
	server.SetupHandler(s, "/", "stable-match", func(c socketio.Conn) {
		fmt.Println("Starting matching process...")
		// start measuring time
		endTrackingFunc := TrackTime()
		algorithm.StableMatch(UnmarriedMen, Women)
		// stop meassuring time
		duration := endTrackingFunc()
		// emit duration of time
		c.Emit("stable-match-duration", duration.String())

		log.Println(fmt.Sprintf("%s took %s", "stable matching", duration))

		if Verbose {
			// print all matches after the stable matching completed
			fmt.Println("Matches:")
			fmt.Println("W <-> M")
			fmt.Println("=======")
			for _, w := range Women {
				fmt.Printf("%d <-> %d\n", w.ID, w.Husband.ID)
			}
		}

	})

	// continue running so that the server won't shut down until signaled
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
}
