package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
)

// Table
type Table struct {
	ID                string       `json:"id"`
	Capacity          int          `json:"capacity"`
	PlayersID         []string     `json:"players_id"`
	Bet               int64        `json:"bet"`
	Private           bool         `json:"private"`
	Match             *Match       `json:"match"`
	PlayersForRematch []string     `json:"players_for_rematch"`
	MatchResult       *MatchResult `json:"match_result,omitempty"`
	GameType          int          `json:"game_type"`
	UpToPoints        int          `json:"up_to_points"`

	room *Room
}

func newTable(room *Room, capacity int, playersID []string, bet int64, private bool, gameType int, upToPoints int) *Table {
	return &Table{
		ID:                fmt.Sprintf("%x", rand.Int()),
		Capacity:          capacity,
		PlayersID:         playersID,
		Bet:               bet,
		Private:           private,
		PlayersForRematch: []string{},
		MatchResult:       nil,
		GameType:          gameType,
		UpToPoints:        upToPoints,
		room:              room,
	}
}

func (table *Table) forwardToMatchAction(action *Action) {
	log.Println("u match stigo action:", action.name)
	switch action.name {
	case "end_match":
		if m := table.Match; m != nil {
			m.endAction(action)
		}
		table.unregisterMatch()
	case "turn":
		if m := table.Match; m != nil {
			m.turnAction(action)
		}
	}
}

func (table *Table) unregisterMatch() {
	log.Println("Unregister match")
	table.room.mu.Lock()
	table.Match = nil
	table.room.mu.Unlock()

	table.room.chBroadcastAll <- table.room.info()
}

func (table *Table) joinAction(action *Action) {
	table.room.mu.Lock()

	joinPlayer := table.room.players[action.fromPlayerID]
	table.PlayersForRematch = []string{}
	if !table.sitPlayer(joinPlayer.ID) {
		log.Println("Can't join to table")
		table.room.mu.Unlock()
		return
	}

	if table.isFull() {
		m := newMatch(table)
		table.Match = m
		table.room.mu.Unlock()
		m.takeInitialBet()
		m.run()
		log.Println("Created new match")
	}

	table.room.chBroadcastAll <- table.room.info()

}

func (table *Table) leave(leavePlayerID string) {
	foundInTable := false
	otherPlayersID := otherPlayersID(table.PlayersID, leavePlayerID)
	for _, playerID := range table.PlayersID {
		if playerID == leavePlayerID {
			foundInTable = true
		}
	}

	if !foundInTable {
		log.Println("leave: Player not found in table!")
		return
	}

	if m := table.Match; m != nil {
		m.leave(leavePlayerID)
		table.unregisterMatch()
	}

	if p := table.room.players[leavePlayerID]; p != nil && p.TableID != nil {
		if *p.TableID == table.ID {
			p.TableID = nil
		}
	}

	table.PlayersID = otherPlayersID
}

func (table *Table) leaveAction(action *Action) {
	table.leave(action.fromPlayerID)
	table.room.chBroadcast <- Broadcast{playersID: table.PlayersID, message: action.message}
}

func (table *Table) rematchAction(action *Action) {
	for _, playerID := range table.PlayersForRematch {
		if playerID == action.fromPlayerID {
			// already ready for rematch
			return
		}
	}
	table.PlayersForRematch = append(table.PlayersForRematch, action.fromPlayerID)
	otherPlayersID := []string{}
	for _, playerID := range table.PlayersID {
		if playerID != action.fromPlayerID {
			otherPlayersID = append(otherPlayersID, playerID)
		}
	}

	if len(table.PlayersForRematch) == table.Capacity {
		table.room.mu.Lock()
		m := newMatch(table)
		table.Match = m
		table.room.mu.Unlock()
		m.takeInitialBet()
		m.table.PlayersForRematch = []string{}

		table.room.chBroadcastAll <- table.room.info()
		m.run()
	}

	table.room.chBroadcast <- Broadcast{playersID: otherPlayersID, message: action.message}

}

func (table *Table) isFull() bool {
	return len(table.PlayersID) >= table.Capacity
}

func (table *Table) sitPlayer(playerID string) bool {
	if table.isFull() {
		log.Println("Ooops table is full!")
		return false
	}

	for _, tablePlayerID := range table.PlayersID {
		if playerID == tablePlayerID {
			log.Println("Ooops, already is for table")
			return false
		}
	}
	table.PlayersID = append(table.PlayersID, playerID)
	table.room.players[playerID].TableID = &table.ID
	log.Printf("Player %v sit on table %v", playerID, table.ID)
	return true
}

func (table *Table) notifyMatchWillStart(period int) {
	js, err := json.Marshal(struct {
		MsgFunc string `json:"msg_func"`
		TableID string `json:"table_id"`
		Period  int    `json:"period"`
	}{
		MsgFunc: "table_match_will_start",
		TableID: table.ID,
		Period:  period,
	})

	if err != nil {
		log.Println(err)
	}

	table.room.chBroadcast <- Broadcast{playersID: table.PlayersID, message: js}
}
