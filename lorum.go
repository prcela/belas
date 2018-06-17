package main

type LorumGame struct {
	State int
}

func (lorum *LorumGame) run() CardGameStep {
	return CardGameStep{}
}

func (lorum *LorumGame) state() int {
	return lorum.State
}

func (lorum *LorumGame) groups() []*CardGroup {
	return []*CardGroup{}
}

func (lorum *LorumGame) nextStep() CardGameStep {
	return CardGameStep{}
}

func (lorum *LorumGame) onPlayerAction(action *Action) CardGameStep {
	return CardGameStep{}
}

func (lorum *LorumGame) group(ID string) *CardGroup {
	return nil
}
