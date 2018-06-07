package main

import (
	"encoding/json"
	"log"
	"time"
)

const (
	BelaStateInit        = 0
	BelaStateDealed      = 1
	BelaStateCall        = 2
	BelaStatePickedTalon = 3
	BelaStatePlay        = 4
)

type BelaGame struct {
	State               int          `json:"state"`
	InitialGroup        *CardGroup   `json:"initial_group"`
	CenterGroup         *CardGroup   `json:"center_group"`
	HandGroups          []*CardGroup `json:"hand_groups"`
	TalonGroups         []*CardGroup `json:"talon_groups"`
	WinGroups           []*CardGroup `json:"win_groups"`
	IdxPlayerOnTurn     int          `json:"idx_player_on_turn"`
	IdxPlayerStartRound int          `json:"idx_player_start_round"`
	IdxPlayerCalled     *int         `json:"idx_player_called"`
}

func (bela *BelaGame) run() CardGameStep {
	bela.InitialGroup = &CardGroup{ID: "Initial"}
	bela.CenterGroup = &CardGroup{ID: "Center", Capacity: 4, Visibility: 2}
	bela.HandGroups = []*CardGroup{
		&CardGroup{ID: "Hand0", Capacity: 8, Visibility: 1},
		&CardGroup{ID: "Hand1", Capacity: 8, Visibility: 1},
		&CardGroup{ID: "Hand2", Capacity: 8, Visibility: 1},
		&CardGroup{ID: "Hand3", Capacity: 8, Visibility: 1}}
	bela.TalonGroups = []*CardGroup{
		&CardGroup{ID: "Talon0", Capacity: 2},
		&CardGroup{ID: "Talon1", Capacity: 2},
		&CardGroup{ID: "Talon2", Capacity: 2},
		&CardGroup{ID: "Talon3", Capacity: 2}}
	bela.WinGroups = []*CardGroup{
		&CardGroup{ID: "Win0"},
		&CardGroup{ID: "Win1"},
	}

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
		WaitDuration:     time.Second,
		Transitions:      []CardTransition{},
		SendCompleteGame: true,
	}

}

func (bela *BelaGame) moveCard(card Card, fromGroup *CardGroup, toGroup *CardGroup) {
	var fromIdx *int
	for idx, c := range fromGroup.Cards {
		if c.Boja == card.Boja && c.Broj == card.Broj {
			fromIdx = &idx
			break
		}
	}
	if fromIdx != nil {
		card := fromGroup.Cards[*fromIdx]
		fromGroup.Cards = append(fromGroup.Cards[:*fromIdx], fromGroup.Cards[*fromIdx+1:]...)
		toGroup.Cards = append(toGroup.Cards, card)
	}
}

func (bela *BelaGame) nextPlayer() {
	bela.IdxPlayerOnTurn = (bela.IdxPlayerOnTurn + 1) % 4
	log.Println("nextPlayer: IdxPlayerOnTurn:", bela.IdxPlayerOnTurn)
}

func (bela *BelaGame) nextStep() CardGameStep {
	log.Println("Bela state: ", bela.State)
	switch bela.State {
	case BelaStateInit:
		return bela.dealStep()
	case BelaStateDealed:
		return bela.callStep()
	case BelaStateCall:
		if bela.IdxPlayerCalled != nil {
			bela.IdxPlayerOnTurn = bela.IdxPlayerStartRound
			return bela.pickTalonStep()
		} else {
			bela.nextPlayer()
			return bela.callStep()
		}
	case BelaStatePickedTalon:
		return bela.playStep()
	case BelaStatePlay:
		bela.nextPlayer()
		return bela.playStep()
	}

	return CardGameStep{}
}

func (bela *BelaGame) onPlayerAction(action *Action) CardGameStep {
	log.Println("onPlayerAction")
	var dic struct {
		Turn string           `json:"turn"`
		Move *CardEnabledMove `json:"enabled_move,omitempty"`
	}
	if err := json.Unmarshal(action.message, &dic); err != nil {
		panic(err)
	}
	step := CardGameStep{
		WaitDuration: 1,
	}
	if dic.Move != nil {
		if dic.Move.ToGroupId != nil {
			fromGroup := bela.group(dic.Move.FromGroupId)
			toGroup := bela.group(*dic.Move.ToGroupId)
			bela.moveCard(dic.Move.Card, fromGroup, toGroup)
			step.Transitions = append(step.Transitions, CardTransition{
				Card:         dic.Move.Card,
				FromGroupId:  dic.Move.FromGroupId,
				ToGroupId:    *dic.Move.ToGroupId,
				ToIdx:        len(toGroup.Cards) - 1,
				WaitDuration: 0,
				Duration:     0.5,
			})
		}
		if bela.State == BelaStateCall {
			bela.IdxPlayerCalled = &bela.IdxPlayerOnTurn
		}
	}
	return step
}

func (bela *BelaGame) dealStep() CardGameStep {
	log.Println("dealStep")
	step := CardGameStep{}
	for idxGroup, group := range bela.HandGroups {
		for i := 0; i < 6; i++ {
			fromIdx := len(bela.InitialGroup.Cards) - 1
			toIdx := len(group.Cards)
			card := bela.InitialGroup.Cards[fromIdx]
			step.Transitions = append(step.Transitions, CardTransition{
				Card:         card,
				FromGroupId:  bela.InitialGroup.ID,
				ToGroupId:    group.ID,
				ToIdx:        toIdx,
				WaitDuration: 0.2*float32(i) + 1.2*float32(idxGroup),
				Duration:     0.5,
			})
			bela.moveCard(card, bela.InitialGroup, group)
		}
	}
	for idxGroup, group := range bela.TalonGroups {
		for i := 0; i < 2; i++ {
			fromIdx := len(bela.InitialGroup.Cards) - 1
			toIdx := len(group.Cards)
			card := bela.InitialGroup.Cards[fromIdx]
			step.Transitions = append(step.Transitions, CardTransition{
				Card:         card,
				FromGroupId:  bela.InitialGroup.ID,
				ToGroupId:    group.ID,
				ToIdx:        toIdx,
				WaitDuration: 5 + 0.2*float32(i) + 1.2*float32(idxGroup),
				Duration:     0.5,
			})
			bela.moveCard(card, bela.InitialGroup, group)
		}
	}
	step.WaitDuration = 10 * time.Second
	bela.State = BelaStateDealed
	return step
}

func (bela *BelaGame) callStep() CardGameStep {
	log.Println("callStep")
	enabledMoves := []CardEnabledMove{}
	fromGroup := bela.HandGroups[bela.IdxPlayerOnTurn]
	log.Println(fromGroup.Cards)
	for _, card := range fromGroup.Cards {
		enabledMoves = append(enabledMoves, CardEnabledMove{FromGroupId: fromGroup.ID, Card: card})
	}
	m := map[int][]CardEnabledMove{bela.IdxPlayerOnTurn: enabledMoves}
	log.Println("enabledMoves:", m)
	bela.State = BelaStateCall
	return CardGameStep{
		EnabledMoves: m,
	}
}

func (bela *BelaGame) pickTalonStep() CardGameStep {
	log.Println("pickTalonStep")
	step := CardGameStep{}
	for idxGroup, group := range bela.TalonGroups {
		for i := 0; i < 2; i++ {
			handGroup := bela.HandGroups[idxGroup]
			fromIdx := len(group.Cards) - 1
			toIdx := len(handGroup.Cards)
			card := group.Cards[fromIdx]
			step.Transitions = append(step.Transitions, CardTransition{
				Card:         card,
				FromGroupId:  group.ID,
				ToGroupId:    handGroup.ID,
				ToIdx:        toIdx,
				WaitDuration: 0.2*float32(i) + 1.2*float32(idxGroup),
				Duration:     0.5,
			})
			bela.moveCard(card, group, handGroup)
		}
	}
	step.WaitDuration = 6 * time.Second
	bela.State = BelaStatePickedTalon
	return step
}

func (bela *BelaGame) playStep() CardGameStep {
	enabledMoves := []CardEnabledMove{}
	fromGroup := bela.HandGroups[bela.IdxPlayerOnTurn]
	log.Println(fromGroup.Cards)
	for _, card := range fromGroup.Cards {
		enabledMoves = append(enabledMoves, CardEnabledMove{FromGroupId: fromGroup.ID, Card: card, ToGroupId: &bela.CenterGroup.ID})
	}
	m := map[int][]CardEnabledMove{bela.IdxPlayerOnTurn: enabledMoves}
	log.Println("enabledMoves:", m)
	bela.State = BelaStatePlay
	return CardGameStep{
		EnabledMoves: m,
		WaitDuration: 60 * time.Second,
	}
}

func (bela *BelaGame) groups() []*CardGroup {
	result := []*CardGroup{bela.InitialGroup, bela.CenterGroup}
	for _, group := range bela.HandGroups {
		result = append(result, group)
	}
	for _, group := range bela.TalonGroups {
		result = append(result, group)
	}
	for _, group := range bela.WinGroups {
		result = append(result, group)
	}
	return result
}

func (bela *BelaGame) group(ID string) *CardGroup {
	for _, g := range bela.groups() {
		if g.ID == ID {
			return g
		}
	}
	return nil
}
