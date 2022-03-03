/*

Performs stable matching between a group of `n` men and women.
Preference ranking is in ascending order, which means lower value has a higher preference.

*/
package main

import (
	"container/list"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	"github.com/Hofled/stable-matching-go/types"
)

var (
	UnmarriedMen *list.List
	Women        []*types.Woman
)

func init() {
	UnmarriedMen = list.New()

	groupSizeArg := os.Args[1]
	fmt.Printf("Group size: %s\n", groupSizeArg)
	// get each group's size
	groupSize, err := strconv.Atoi(groupSizeArg)
	if err != nil {
		log.Fatal(err)
	}

	Women = make([]*types.Woman, groupSize)

	// generate men
	for i := 0; i < groupSize; i++ {
		man := types.NewMan(i, groupSize)
		usedRankingsMap := make(map[int]bool, groupSize)
		// generate random preferences for the i'th man
		for j := 0; j < groupSize; j++ {
			man.Preferences[j] = UniqueRandIntInRange(usedRankingsMap, groupSize)
		}
		// add newly created man to the men list
		UnmarriedMen.PushBack(man)
	}

	// generate women
	for i := 0; i < groupSize; i++ {
		// generate preferences for the i'th woman
		woman := types.NewWoman(i, groupSize)
		usedRankingsMap := make(map[int]bool, groupSize)
		// generate random preferences for the i'th man
		for j := 0; j < groupSize; j++ {
			woman.Preferences[j] = UniqueRandIntInRange(usedRankingsMap, groupSize)
		}
		// add newly created man to the men list
		Women[i] = woman
	}
}

func UniqueRandIntInRange(usedValsMap map[int]bool, max int) int {
	val := rand.Intn(max)
	for usedValsMap[val] {
		val = rand.Intn(max)
	}
	// insert used val into the used vals map
	usedValsMap[val] = true
	return val
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

func main() {
	fmt.Println("Men:")
	for m := UnmarriedMen.Front(); m != nil; m = m.Next() {
		fmt.Printf("m.Value: %v\n", m.Value)
	}

	fmt.Println("Women:")
	for _, w := range Women {
		fmt.Printf("w.Value: %v\n", w)
	}

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

	// print all matches after the stable matching completed
	fmt.Println("Matches:")
	fmt.Println("W <-> M")
	fmt.Println("=======")
	for _, w := range Women {
		fmt.Printf("%d <-> %d\n", w.ID, w.Husband.ID)
	}
}