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
	CenterGroups        []*CardGroup `json:"center_groups"`
	HandGroups          []*CardGroup `json:"hand_groups"`
	TalonGroups         []*CardGroup `json:"talon_groups"`
	WinGroups           []*CardGroup `json:"win_groups"`
	IdxPlayerDealed     int          `json:"idx_player_dealed"`
	IdxPlayerOnTurn     int          `json:"idx_player_on_turn"`
	IdxPlayerStartRound int          `json:"idx_player_start_round"`
	IdxPlayerCalled     *int         `json:"idx_player_called"`
	Adut                *string      `json:"adut"`
}

func (bela *BelaGame) state() int {
	return bela.State
}

func (bela *BelaGame) run() CardGameStep {
	bela.InitialGroup = &CardGroup{ID: "Initial"}
	bela.HandGroups = []*CardGroup{
		&CardGroup{ID: "Hand0", Visibility: 1},
		&CardGroup{ID: "Hand1", Visibility: 1},
		&CardGroup{ID: "Hand2", Visibility: 1},
		&CardGroup{ID: "Hand3", Visibility: 1}}
	bela.TalonGroups = []*CardGroup{
		&CardGroup{ID: "Talon0"},
		&CardGroup{ID: "Talon1"},
		&CardGroup{ID: "Talon2"},
		&CardGroup{ID: "Talon3"}}
	bela.WinGroups = []*CardGroup{
		&CardGroup{ID: "Win0"},
		&CardGroup{ID: "Win1"},
	}
	bela.CenterGroups = []*CardGroup{
		&CardGroup{ID: "Center0", Visibility: 2},
		&CardGroup{ID: "Center1", Visibility: 2},
		&CardGroup{ID: "Center2", Visibility: 2},
		&CardGroup{ID: "Center3", Visibility: 2},
	}

	cards := []Card{}
	for _, boja := range []string{"zir", "bundeva", "list", "srce"} {
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
			return bela.callStep()
		}
	case BelaStatePickedTalon:
		return bela.playStep()
	case BelaStatePlay:
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
		WaitDuration: 1 * time.Second,
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
				ToTop:        true,
				WaitDuration: 0,
				Duration:     0.5,
			})
		}
		if bela.State == BelaStateCall {
			bela.IdxPlayerCalled = &bela.IdxPlayerOnTurn
			bela.Adut = &dic.Move.Card.Boja
			step.CardGameEvent = &CardGameEvent{
				Category: "Player",
				Action:   "Call",
				Label:    *bela.Adut,
				Value:    *bela.IdxPlayerCalled,
			}
			log.Println(step)
		}
		bela.nextPlayer()
	}
	return step
}

func (bela *BelaGame) dealStep() CardGameStep {
	log.Println("dealStep")
	step := CardGameStep{}
	for idxGroup, group := range bela.HandGroups {
		for i := 0; i < 6; i++ {
			fromIdx := len(bela.InitialGroup.Cards) - 1
			card := bela.InitialGroup.Cards[fromIdx]
			step.Transitions = append(step.Transitions, CardTransition{
				Card:         card,
				FromGroupId:  bela.InitialGroup.ID,
				ToGroupId:    group.ID,
				ToTop:        true,
				WaitDuration: 0.2*float32(i) + 1.2*float32(idxGroup),
				Duration:     0.5,
			})
			bela.moveCard(card, bela.InitialGroup, group)
		}
	}
	for idxGroup, group := range bela.TalonGroups {
		for i := 0; i < 2; i++ {
			fromIdx := len(bela.InitialGroup.Cards) - 1
			card := bela.InitialGroup.Cards[fromIdx]
			step.Transitions = append(step.Transitions, CardTransition{
				Card:         card,
				FromGroupId:  bela.InitialGroup.ID,
				ToGroupId:    group.ID,
				ToTop:        true,
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
			card := group.Cards[fromIdx]
			step.Transitions = append(step.Transitions, CardTransition{
				Card:         card,
				FromGroupId:  group.ID,
				ToGroupId:    handGroup.ID,
				ToTop:        true,
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

func (bela *BelaGame) cardStrength(card Card) int {
	if card.Boja == *bela.Adut {
		switch card.Broj {
		case 7:
			return 15
		case 8:
			return 16
		case 9:
			return 21
		case 10:
			return 19
		case 11:
			return 22
		case 12:
			return 17
		case 13:
			return 18
		case 14:
			return 20
		}
	}
	return card.Broj
}

func (bela *BelaGame) playStep() CardGameStep {

	centerCards := []Card{}
	for i := 0; i < 4; i++ {
		group := bela.CenterGroups[(bela.IdxPlayerStartRound+i)%4]
		centerCards = append(centerCards, group.Cards...)
	}

	// ako su sve karte pale
	if len(centerCards) == 4 {
		log.Println("sve 4 karte su pale")
		step := CardGameStep{}
		toGroup := bela.WinGroups[0]
		for _, group := range bela.CenterGroups {
			for _, card := range group.Cards {
				step.Transitions = append(step.Transitions, CardTransition{
					Card:         card,
					FromGroupId:  group.ID,
					ToGroupId:    toGroup.ID,
					ToTop:        true,
					WaitDuration: 0,
					Duration:     0.5,
				})
				bela.moveCard(card, group, toGroup)
			}
		}
		step.WaitDuration = 1 * time.Second
		return step
	}

	enabledMoves := []CardEnabledMove{}
	fromGroup := bela.HandGroups[bela.IdxPlayerOnTurn]
	toGroup := bela.CenterGroups[bela.IdxPlayerOnTurn]
	log.Println(fromGroup.Cards)

	// ako je već karta u centru
	if len(centerCards) > 0 {
		cardFirstInCenter := centerCards[0]
		strongestInCenter := cardFirstInCenter
		presjeceno := false
		for _, card := range centerCards[1:] {
			if bela.cardStrength(card) > bela.cardStrength(strongestInCenter) {
				if card.Boja == *bela.Adut {
					if *bela.Adut != cardFirstInCenter.Boja {
						presjeceno = true
					}
				}
				strongestInCenter = card
			}
		}

		if presjeceno {
			// moraš poštivati boju, ne treba jača
			if len(enabledMoves) == 0 {
				for _, card := range fromGroup.Cards {
					if card.Boja == cardFirstInCenter.Boja {
						enabledMoves = append(enabledMoves, CardEnabledMove{FromGroupId: fromGroup.ID, Card: card, ToGroupId: &toGroup.ID})
					}
				}
			}
			// ako nema u istoj boji, moraš aduta
			if len(enabledMoves) == 0 {
				for _, card := range fromGroup.Cards {
					if card.Boja == *bela.Adut {
						enabledMoves = append(enabledMoves, CardEnabledMove{FromGroupId: fromGroup.ID, Card: card, ToGroupId: &toGroup.ID})
					}
				}
			}
		} else {
			// u istoj boji moraš jaču odigrati
			for _, card := range fromGroup.Cards {
				if card.Boja == cardFirstInCenter.Boja && bela.cardStrength(card) > bela.cardStrength(strongestInCenter) {
					enabledMoves = append(enabledMoves, CardEnabledMove{FromGroupId: fromGroup.ID, Card: card, ToGroupId: &toGroup.ID})
				}
			}
			// ako nema jaču u istoj boji, probaj slabiju u istoj boji
			if len(enabledMoves) == 0 {
				for _, card := range fromGroup.Cards {
					if card.Boja == cardFirstInCenter.Boja {
						enabledMoves = append(enabledMoves, CardEnabledMove{FromGroupId: fromGroup.ID, Card: card, ToGroupId: &toGroup.ID})
					}
				}
			}
			// ako nema u istoj boji, moraš aduta
			if len(enabledMoves) == 0 {
				for _, card := range fromGroup.Cards {
					if card.Boja == *bela.Adut {
						enabledMoves = append(enabledMoves, CardEnabledMove{FromGroupId: fromGroup.ID, Card: card, ToGroupId: &toGroup.ID})
					}
				}
			}
		}
	}

	// ako nemaš još ništa od navedenog, baci bilo šta
	if len(enabledMoves) == 0 {
		for _, card := range fromGroup.Cards {
			enabledMoves = append(enabledMoves, CardEnabledMove{FromGroupId: fromGroup.ID, Card: card, ToGroupId: &toGroup.ID})
		}
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
	result := []*CardGroup{bela.InitialGroup}
	result = append(result, bela.HandGroups...)
	result = append(result, bela.CenterGroups...)
	result = append(result, bela.TalonGroups...)
	result = append(result, bela.WinGroups...)
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
