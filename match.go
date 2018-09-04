package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"log"
	"math/rand"
	"time"
)

// Match ...
type Match struct {
	ID                  string   `json:"id"`
	PlayersID           []string `json:"players_id"`
	IndexOfPlayerOnTurn int      `json:"ìndex_of_player_on_turn"`
	TurnDuration        int      `json:"turn_duration"`

	table         *Table
	cardGame      CardGame
	WaitDurations []time.Duration
	chPlayerTurn  chan *Action

	chWaitNextPTurn chan *Action
	chLeave         chan string
}

func newMatch(table *Table) *Match {
	return &Match{
		ID:                  fmt.Sprintf("%x", rand.Int()),
		PlayersID:           table.PlayersID,
		IndexOfPlayerOnTurn: 0,
		TurnDuration:        60,
		table:               table,
		cardGame:            table.room.newCardGame(),
		WaitDurations:       []time.Duration{0, 0, 0, 0},
		chWaitNextPTurn:     make(chan *Action),
		chPlayerTurn:        make(chan *Action),
	}
}

func (m *Match) run() {
	if m == nil {
		return
	}
	m.notifyStarted()
	nextStep := m.cardGame.run()
	var ticker *time.Ticker

	process := func(nextStep CardGameStep) {
		if ticker != nil {
			log.Println("ticker stopped")
			ticker.Stop()
		}

		if nextStep.WaitDuration > 0 {
			log.Println("NewTicker:", nextStep.WaitDuration)
			ticker = time.NewTicker(nextStep.WaitDuration)
		}

		if nextStep.SendCompleteGame {
			msgNum := newMsgNum()
			js, err := json.Marshal(struct {
				MsgFunc  string   `json:"msg_func"`
				MsgNum   int32    `json:"msg_num"`
				CardGame CardGame `json:"game"`
			}{
				MsgFunc:  "game",
				MsgNum:   msgNum,
				CardGame: m.cardGame,
			})
			if err != nil {
				log.Println(err)
			}
			m.table.room.chBroadcast <- Broadcast{
				playersID: m.table.PlayersID,
				message:   js,
				msgNum:    msgNum,
			}
		}

		msgNum := newMsgNum()
		js, err := json.Marshal(struct {
			MsgFunc string       `json:"msg_func"`
			State   int          `json:"state"`
			Step    CardGameStep `json:"step"`
			MsgNum  int32        `json:"msg_num"`
		}{
			MsgFunc: "step",
			State:   m.cardGame.state(),
			Step:    nextStep,
			MsgNum:  msgNum,
		})

		if err != nil {
			log.Println(err)
		}
		m.table.room.chBroadcast <- Broadcast{
			playersID: m.table.PlayersID,
			message:   js,
			msgNum:    msgNum,
		}
	}

	process(nextStep)

	for {
		select {
		case <-ticker.C:
			log.Println("match.run: ticker.C")
			nextStep := m.cardGame.nextStep()
			process(nextStep)
		case action := <-m.chPlayerTurn:
			log.Println("match.run: chPlayerTurn")
			nextStep := m.cardGame.onPlayerAction(action)
			process(nextStep)
		}
	}
}

func (m *Match) takeInitialBet() {

	db, s := m.table.room.GetDatabaseSessionCopy()
	defer s.Close()

	m.table.room.mu.Lock()

	for _, playerID := range m.PlayersID {
		if player := m.table.room.players[playerID]; player != nil {
			player.Diamonds -= m.table.Bet
			change := bson.M{"$set": bson.M{
				"diamonds": player.Diamonds,
			}}
			db.C("players").Update(bson.M{"_id": player.ID}, change)
			log.Println("Player diamonds set to: ", player.Diamonds)
		}
	}
	m.table.room.mu.Unlock()

}

func (m *Match) notifyStarted() {
	log.Println("Notify started")
	msgNum := newMsgNum()
	js, err := json.Marshal(struct {
		MsgFunc string `json:"msg_func"`
		Table   *Table `json:"table"`
		MsgNum  int32  `json:"msg_num"`
	}{
		MsgFunc: "match_started",
		Table:   m.table,
		MsgNum:  msgNum,
	})

	if err != nil {
		log.Println(err)
	}
	m.table.room.chBroadcast <- Broadcast{playersID: m.PlayersID, message: js, msgNum: msgNum}
}

func (m *Match) leave(leavePlayerID string) {
	log.Printf("Player %v left the match.\n", leavePlayerID)

	foundInMatch := false
	for _, playerID := range m.PlayersID {
		if playerID == leavePlayerID {
			foundInMatch = true
		}
	}

	if !foundInMatch {
		return
	}

	db, s := m.table.room.GetDatabaseSessionCopy()
	defer s.Close()

	m.table.room.mu.Lock()
	m.table.MatchResult = newMatchResult(m.table.PlayersID, []int{0, 0, 0, 0}, m.WaitDurations)
	for idxPlayer, playerID := range m.PlayersID {
		if player := m.table.room.players[playerID]; player != nil {
			if player.ID != leavePlayerID {
				// this player wins // FIXME!!!!
				m.table.MatchResult.WinnerID = playerID
				m.table.MatchResult.TotalWinnerID = playerID
				m.table.MatchResult.Scores[idxPlayer] = 1

				player.Diamonds += 2 * m.table.Bet
				change := bson.M{"$set": bson.M{
					"diamonds": player.Diamonds,
				}}
				db.C("players").Update(bson.M{"_id": player.ID}, change)
				log.Println("Player diamonds set to: ", player.Diamonds)

			}
		}
	}
	m.PlayersID = otherPlayersID(m.PlayersID, leavePlayerID)
	m.table.room.mu.Unlock()

}

func (m *Match) endAction(action *Action) {
	var dic struct {
		Scores []int `json:"scores"`
	}
	if err := json.Unmarshal(action.message, &dic); err != nil {
		panic(err)
	}
	log.Println(dic)

	otherPlayersInMatch := []string{}
	m.table.MatchResult = newMatchResult(m.table.PlayersID, dic.Scores, m.WaitDurations)
	log.Println("m.table.winnerID winner: ", m.table.MatchResult.WinnerID)
	m.chWaitNextPTurn <- action

	db, s := m.table.room.GetDatabaseSessionCopy()
	defer s.Close()

	m.table.room.mu.Lock()

	for _, playerID := range m.PlayersID {
		if player := m.table.room.players[playerID]; player != nil {
			if player.ID == m.table.MatchResult.WinnerID {
				player.Diamonds += 2 * m.table.Bet
			} else if m.table.MatchResult.WinnerID == "drawn" {
				player.Diamonds += m.table.Bet
			}
			change := bson.M{"$set": bson.M{
				"diamonds": player.Diamonds,
			}}
			db.C("players").Update(bson.M{"_id": player.ID}, change)
			log.Println("Player diamonds set to: ", player.Diamonds)
		}
		if action.fromPlayerID != playerID {
			otherPlayersInMatch = append(otherPlayersInMatch, playerID)
		}
	}

	m.table.room.mu.Unlock()

	msgNum := newMsgNum()
	js, err := json.Marshal(struct {
		MsgFunc string `json:"msg_func"`
		Table   *Table `json:"table"`
		MsgNum  int32  `json:"msg_num"`
	}{
		MsgFunc: "end_match",
		Table:   m.table,
		MsgNum:  msgNum,
	})

	if err != nil {
		log.Println(err)
	}

	m.table.room.chBroadcast <- Broadcast{playersID: m.table.PlayersID, message: js, msgNum: msgNum}
}

func (m *Match) turnAction(action *Action) {

	otherPlayersInMatch := otherPlayersID(m.PlayersID, action.fromPlayerID)
	m.table.room.chBroadcast <- Broadcast{playersID: otherPlayersInMatch, message: action.message, msgNum: action.msgNum}

	m.chPlayerTurn <- action
}
