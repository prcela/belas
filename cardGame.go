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
	FromIdx      int     `json:"from_idx"`
	ToGroupId    string  `json:"to_group_id"`
	ToIdx        int     `json:"to_idx"`
	WaitDuration float32 `json:"wait_duration"`
	Duration     float32 `json:"duration"`
}

type CardEnabledMove struct {
	Card      Card    `json:"card"`
	ToGroupId *string `json:"to_group_id"`
}

type CardGameStep struct {
	WaitDuration     time.Duration
	Transitions      []CardTransition
	EnabledMoves     map[int][]CardEnabledMove
	SendCompleteGame bool
}

type CardGame interface {
	run() CardGameStep
	nextStep() CardGameStep
	onPlayerAction(action *Action) CardGameStep
	groups() []*CardGroup
}

type CardGroup struct {
	id    string `json:"id"`
	Cards []Card `json:"cards"`
}

func (group *CardGroup) shuffle() {
	perm := rand.Perm(len(group.Cards))
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
