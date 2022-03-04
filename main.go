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
	"math/rand"
	"sync"
	"time"

	"github.com/Hofled/stable-matching-go/channels"
	"github.com/Hofled/stable-matching-go/types"
	"github.com/gosuri/uiprogress"
)

const default_group_size = 5

var (
	UnmarriedMen *list.List
	Women        []*types.Woman
	Verbose      bool
	GroupSize    int
)

func init() {
	flag.BoolVar(&Verbose, "v", false, "Whether to print the generated groups info")
	flag.IntVar(&GroupSize, "n", default_group_size, "The group size for each gender")
}

// returns a pointer to the element in the unmarried men list that needs to be removed (newly married),
// if non needs to be removed, returns nil
func Propose(man *list.Element, woman *types.Woman) *list.Element {
	castedMan := man.Value.(*types.Man)
	// if woman is unmarried, accept the proposal
	if woman.Husband == nil {
		woman.Husband = castedMan
		// remove the new husband
		return man
	} else {
		// if the man proposing has a higher ranking than the current husband, marry him instead
		if woman.Preferences[castedMan.ID] < woman.Preferences[woman.Husband.ID] {
			// add the current husband to the unmarried man linked list
			UnmarriedMen.PushBack(woman.Husband)
			woman.Husband = castedMan
			// remove the new husband
			return man
		}
	}
	castedMan.ProposeIndex++
	return nil
}

func TrackTime() func() time.Duration {
	start := time.Now()
	// returns the duration since we started tracking time
	return func() time.Duration {
		return time.Since(start)
	}
}

func generateGroups() {
	UnmarriedMen = list.New()

	fmt.Printf("Group size: %d\n", GroupSize)

	Women = make([]*types.Woman, GroupSize)

	var menGeneration sync.WaitGroup
	menGenerationChan := make(chan interface{})
	var womenGeneration sync.WaitGroup
	womenGenerationChan := make(chan interface{})

	mergedMenChan := channels.MergedChannel{OnClosed: func() {}, ReceivingChan: menGenerationChan}
	mergedWomenChan := channels.MergedChannel{OnClosed: func() {}, ReceivingChan: womenGenerationChan}

	genFinished := channels.Merge(mergedMenChan, mergedWomenChan)

	fmt.Println("Generating preferences for men & women:")

	uiprogress.Start()

	menGenBar := uiprogress.AddBar(GroupSize).AppendCompleted().PrependFunc(generationPrependFunc("Men gen"))
	menGeneration.Add(GroupSize)
	// generate men in concurrent routines
	for i := 0; i < GroupSize; i++ {
		go func(index int) {
			man := types.NewMan(index, GroupSize)
			// generate random preferences for the i'th man
			man.Preferences = rand.Perm(GroupSize)
			// add newly created man to the men list
			UnmarriedMen.PushBack(man)
			menGenBar.Incr()
			menGeneration.Done()
		}(i)
	}

	womenGenBar := uiprogress.AddBar(GroupSize).AppendCompleted().PrependFunc(generationPrependFunc("Women gen"))
	womenGeneration.Add(GroupSize)
	// generate women
	for i := 0; i < GroupSize; i++ {
		go func(index int) {
			// generate preferences for the i'th woman
			woman := types.NewWoman(index, GroupSize)
			// generate random preferences for the i'th man
			woman.Preferences = rand.Perm(GroupSize)
			// add newly created man to the men list
			Women[index] = woman
			womenGenBar.Incr()
			womenGeneration.Done()
		}(i)
	}

	channels.CloseWhenDone(&menGeneration, menGenerationChan)
	channels.CloseWhenDone(&womenGeneration, womenGenerationChan)

	genFinished.Wait()
	uiprogress.Stop()
}

func generationPrependFunc(name string) uiprogress.DecoratorFunc {
	return func(b *uiprogress.Bar) string {
		return fmt.Sprintf("[%s]: %d/%d", name, b.Current(), b.Total)
	}
}

func main() {
	flag.Parse()

	generateGroups()

	if Verbose {
		fmt.Println("Men:")
		for m := UnmarriedMen.Front(); m != nil; m = m.Next() {
			fmt.Println(m.Value)
		}

		fmt.Println("Women:")
		for _, w := range Women {
			fmt.Println(w)
		}
	}

	fmt.Println("Starting matching process...")

	// start measuring time
	endTrackingFunc := TrackTime()

	m := UnmarriedMen.Front()
	for m != nil {
		castedMan := m.Value.(*types.Man)
		womanID := castedMan.Preferences[castedMan.ProposeIndex]
		newlyMarriedMan := Propose(m, Women[womanID])
		if newlyMarriedMan != nil {
			m = m.Next()
			UnmarriedMen.Remove(newlyMarriedMan)
		}
	}

	// stop meassuring time
	duration := endTrackingFunc()

	if Verbose {
		// print all matches after the stable matching completed
		fmt.Println("Matches:")
		fmt.Println("W <-> M")
		fmt.Println("=======")
		for _, w := range Women {
			fmt.Printf("%d <-> %d\n", w.ID, w.Husband.ID)
		}
	}

	log.Println(fmt.Sprintf("%s took %s", "stable matching", duration))
}
