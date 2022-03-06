/*

Performs stable matching between a group of `n` men and women.
Preference ranking is in descending order, which means higher value means a higher preference.

*/
package main

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Hofled/stable-matching-go/internal/app/algorithm"
	"github.com/Hofled/stable-matching-go/internal/app/types"
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

	// setup handlers
	UnmarriedMen, _, Women = algorithm.GenerateGroups(GroupSize, Verbose)
	fmt.Println("Starting matching process...")
	// start measuring time
	endTrackingFunc := TrackTime()
	algorithm.StableMatch(UnmarriedMen, Women)
	// stop meassuring time
	duration := endTrackingFunc()

	log.Println(fmt.Sprintf("%s took %s", "stable matching", duration))

	if Verbose {
		// print all matches after the stable matching completed
		fmt.Println("Matches:")
		fmt.Println("W <=> M")
		fmt.Println("=======")
		for _, w := range Women {
			fmt.Printf("%d <=> %d\n", w.ID, w.Husband.ID)
		}
	}
}
