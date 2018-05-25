package main

import (
	"log"
	"time"
)

const (
	BelaStateInit   = 0
	BelaStateDealed = 1
	BelaStateCall   = 2
)

type BelaGame struct {
	State               int
	InitialGroup        CardGroup
	HandGroups          []CardGroup
	TalonGroups         []CardGroup
	IdxPlayerOnTurn     int
	IdxPlayerStartRound int
}

func (bela *BelaGame) run() CardGameStep {
	bela.InitialGroup = CardGroup{id: "Initial"}
	bela.HandGroups = []CardGroup{
		CardGroup{id: "Hand0"},
		CardGroup{id: "Hand1"},
		CardGroup{id: "Hand2"},
		CardGroup{id: "Hand3"}}
	bela.TalonGroups = []CardGroup{
		CardGroup{id: "Talon0"},
		CardGroup{id: "Talon1"},
		CardGroup{id: "Talon2"},
		CardGroup{id: "Talon3"}}

	cards := []Card{}
	for _, boja := range []string{"žir", "bundeva", "list", "srce"} {
		for _, broj := range []int{7, 8, 9, 10, 11, 12, 13, 14} {
			cards = append(cards, Card{Boja: boja, Broj: broj})
		}
	}

	bela.InitialGroup.Cards = cards
	bela.InitialGroup.shuffle()
	bela.IdxPlayerOnTurn = 0
	bela.State = BelaStateInit

	return CardGameStep{
		WaitDuration: time.Second,
		Transitions:  []CardTransition{},
	}

}

func (bela *BelaGame) moveCard(fromGroup *CardGroup, fromIdx int, toGroup *CardGroup) {
	card := fromGroup.Cards[fromIdx]
	fromGroup.Cards = append(fromGroup.Cards[:fromIdx], fromGroup.Cards[fromIdx+1:]...)
	toGroup.Cards = append(toGroup.Cards, card)
}

func (bela *BelaGame) nextStep() CardGameStep {
	log.Println("Bela state: ", bela.State)
	switch bela.State {
	case BelaStateInit:
		return bela.dealStep()
	case BelaStateDealed:
		return bela.callStep()
	}
	return CardGameStep{}
}

func (bela *BelaGame) onPlayerAction(action *Action) CardGameStep {
	return CardGameStep{}
}

func (bela *BelaGame) dealStep() CardGameStep {
	step := CardGameStep{}
	for idxGroup, group := range bela.HandGroups {
		for i := 0; i < 7; i++ {
			fromIdx := len(bela.InitialGroup.Cards) - 1
			toIdx := len(group.Cards)
			step.Transitions = append(step.Transitions, CardTransition{
				Card:         bela.InitialGroup.Cards[fromIdx],
				FromGroupId:  bela.InitialGroup.id,
				FromIdx:      fromIdx,
				ToGroupId:    group.id,
				ToIdx:        toIdx,
				WaitDuration: 0.2*float32(i) + 1.2*float32(idxGroup),
				Duration:     0.5,
			})
			bela.moveCard(&bela.InitialGroup, fromIdx, &group)
		}
	}
	step.WaitDuration = time.Second
	bela.State = BelaStateDealed
	return step
}

func (bela *BelaGame) callStep() CardGameStep {
	return CardGameStep{}
}

func (bela *BelaGame) groups() []CardGroup {
	result := []CardGroup{bela.InitialGroup}
	for _, group := range bela.HandGroups {
		result = append(result, group)
	}
	for _, group := range bela.TalonGroups {
		result = append(result, group)
	}
	return result
}
