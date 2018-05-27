package main

type LorumGame struct {
}

func (lorum *LorumGame) run() CardGameStep {
	return CardGameStep{}
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
