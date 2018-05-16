package main

type Card struct {
	Boja string `json:"boja"`
	Broj int    `json:"broj"`
}

type CardGame interface {
	allCards() []Card
	run()
	cards(groupName string) []Card
}
