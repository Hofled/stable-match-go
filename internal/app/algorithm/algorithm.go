package algorithm

import (
	"container/list"
	"fmt"
	"math/rand"
	"sync"

	"github.com/Hofled/stable-matching-go/internal/app/channels"
	"github.com/Hofled/stable-matching-go/internal/app/types"
	"github.com/gosuri/uiprogress"
)

type MatchingStep struct {
	HusbandID   int
	WifeID      int
	UnmarriedID int
}

type MatchingHistory struct {
	Steps []MatchingStep
}

// returns a pointer to the element in the unmarried men list that needs to be removed (newly married),
// if non needs to be removed, returns nil
func propose(man *list.Element, unmarriedMen *list.List, woman *types.Woman) *list.Element {
	castedMan := man.Value.(*types.Man)
	// if woman is unmarried, accept the proposal
	if woman.Husband == nil {
		woman.Husband = castedMan
		// remove the new husband
		return man
	} else {
		// if the man proposing has a higher ranking than the current husband, marry him instead
		if woman.Preferences[castedMan.ID] > woman.Preferences[woman.Husband.ID] {
			// add the current husband to the unmarried man linked list
			unmarriedMen.PushBack(woman.Husband)
			woman.Husband = castedMan
			// remove the new husband
			return man
		}
	}
	castedMan.ProposeIndex++
	return nil
}

func generationPrependFunc(name string) uiprogress.DecoratorFunc {
	return func(b *uiprogress.Bar) string {
		return fmt.Sprintf("[%s]: %d/%d", name, b.Current(), b.Total)
	}
}

func GenerateGroups(groupSize int, verbose bool) (unmarriedMen *list.List, unmarriedMenSlice []*types.Man, women []*types.Woman) {
	unmarriedMen = list.New()

	fmt.Printf("Group size: %d\n", groupSize)

	women = make([]*types.Woman, groupSize)
	unmarriedMenSlice = make([]*types.Man, groupSize)

	var menGeneration sync.WaitGroup
	menGenerationChan := make(chan interface{})
	var womenGeneration sync.WaitGroup
	womenGenerationChan := make(chan interface{})

	mergedMenChan := channels.MergedChannel{OnClosed: func() {}, ReceivingChan: menGenerationChan}
	mergedWomenChan := channels.MergedChannel{OnClosed: func() {}, ReceivingChan: womenGenerationChan}

	genFinished := channels.Merge(mergedMenChan, mergedWomenChan)

	fmt.Println("Generating preferences for men & women:")

	p := uiprogress.New()
	p.Start()

	menGenBar := p.AddBar(groupSize).AppendCompleted().PrependFunc(generationPrependFunc("Men gen"))
	menGeneration.Add(groupSize)
	// generate men in concurrent routines
	for i := 0; i < groupSize; i++ {
		go func(index int) {
			man := types.NewMan(index, groupSize)
			// generate random preferences for the i'th man
			man.Preferences = rand.Perm(groupSize)
			// add newly created man to the men list
			unmarriedMen.PushBack(man)
			unmarriedMenSlice[index] = man
			menGenBar.Incr()
			menGeneration.Done()
		}(i)
	}

	womenGenBar := p.AddBar(groupSize).AppendCompleted().PrependFunc(generationPrependFunc("Women gen"))
	womenGeneration.Add(groupSize)
	// generate women
	for i := 0; i < groupSize; i++ {
		go func(index int) {
			// generate preferences for the i'th woman
			woman := types.NewWoman(index, groupSize)
			// generate random preferences for the i'th man
			woman.Preferences = rand.Perm(groupSize)
			// add newly created man to the men list
			women[index] = woman
			womenGenBar.Incr()
			womenGeneration.Done()
		}(i)
	}

	channels.CloseWhenDone(&menGeneration, menGenerationChan)
	channels.CloseWhenDone(&womenGeneration, womenGenerationChan)

	genFinished.Wait()
	p.Stop()

	if verbose {
		fmt.Println("Men:")
		for m := unmarriedMen.Front(); m != nil; m = m.Next() {
			fmt.Println(m.Value)
		}

		fmt.Println("Women:")
		for _, w := range women {
			fmt.Println(w)
		}
	}

	return unmarriedMen, unmarriedMenSlice, women
}

func StableMatch(unmarriedMen *list.List, women []*types.Woman) *MatchingHistory {
	history := MatchingHistory{}

	m := unmarriedMen.Front()
	for m != nil {
		proposingMan := m.Value.(*types.Man)
		womanID := proposingMan.Preferences[proposingMan.ProposeIndex]
		currentHusbandID := -1
		if women[womanID].Husband != nil {
			currentHusbandID = women[womanID].Husband.ID
		}
		newlyMarriedMan := propose(m, unmarriedMen, women[womanID])
		if newlyMarriedMan != nil {
			m = m.Next()
			unmarriedMen.Remove(newlyMarriedMan)
			history.Steps = append(history.Steps, MatchingStep{WifeID: womanID, HusbandID: proposingMan.ID, UnmarriedID: currentHusbandID})
		}
	}

	return &history
}
