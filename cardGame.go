package main

import (
	"log"
	"math/rand"
	"time"
)

type Card struct {
	Boja string `json:"boja"`
	Broj int    `json:"broj"`
}

type CardTransition struct {
	Card         Card    `json:"card"`
	FromGroupId  string  `json:"from_group_id"`
	ToGroupId    string  `json:"to_group_id"`
	ToTop        bool    `json:"to_top"`
	WaitDuration float32 `json:"wait_duration"`
	Duration     float32 `json:"duration"`
}

type CardEnabledMove struct {
	FromGroupId string  `json:"from_group_id"`
	Card        Card    `json:"card"`
	ToGroupId   *string `json:"to_group_id"`
}

type CardGameEvent struct {
	Category string `json:"category"`
	Action   string `json:"action"`
	Label    string `json:"label"`
	Value    int    `json:"value"`
}

type CardGameStep struct {
	WaitDuration     time.Duration
	Transitions      []CardTransition
	EnabledMoves     map[int][]CardEnabledMove
	CardGameEvent    *CardGameEvent
	SendCompleteGame bool
}

type CardGame interface {
	run() CardGameStep
	state() int
	nextStep() CardGameStep
	onPlayerAction(action *Action) CardGameStep
	groups() []*CardGroup
	group(ID string) *CardGroup
}

type CardGroup struct {
	ID         string `json:"id"`
	Cards      []Card `json:"cards"`
	Visibility int    `json:"visibility"` // 0 = hidden, 1 = shown to local player only, 2 = visible
}

func (group *CardGroup) shuffle() {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	perm := r1.Perm(len(group.Cards))
	dest := make([]Card, len(group.Cards))
	copy(dest, group.Cards)
	log.Println("len(group.Cards)=", len(group.Cards))
	log.Println("len(dest)=", len(dest))
	for i, v := range perm {
		log.Printf("i: %v v: %v\n", i, v)
		dest[v] = group.Cards[i]
	}
	group.Cards = dest
}
