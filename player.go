package main

import (
	"encoding/json"
	"github.com/fatih/color"
	"log"
	"time"
)

// Player ...
type Player struct {
	ID         string  `json:"id" bson:"_id,omitempty"`
	Alias      string  `json:"alias" bson:"alias"`
	Diamonds   int64   `json:"diamonds" bson:"diamonds"`
	Retentions []int   `json:"retentions" bson:"retentions"`
	TableID    *string `json:"table_id"`

	room           *Room
	missedMessages []MissedMessage
	toAck          map[int32]bool

	// timeLastConnected    *time.Time
	// timeLastDisconnected *time.Time
	chWaitClient     chan *Client // wait for new client
	waitingForClient bool
}

func newPlayer(room *Room, playerID string) *Player {
	return &Player{
		ID:               playerID,
		Alias:            "Player" + playerID,
		Diamonds:         100,
		Retentions:       []int{},
		missedMessages:   []MissedMessage{},
		toAck:            make(map[int32]bool),
		chWaitClient:     make(chan *Client),
		waitingForClient: false,
	}
}

func (p *Player) waitForClientConnection(room *Room, duration time.Duration) {
	color.Green("Waiting for client connection ...")
	ticker := time.NewTicker(duration)
	p.waitingForClient = true

	defer func() {
		ticker.Stop()
		p.waitingForClient = false
	}()

	for {
		select {
		case <-p.chWaitClient:
			log.Println("Jupi, player arrived!")
			return
		case <-ticker.C:
			// timeout
			log.Println("Timeout in waiting for client connection!")
			room.chUnregisterPlayer <- p
			return
		}
	}
}

func (room *Room) resendMissedMessages(playerID string, missedMessages []MissedMessage) {
	color.Blue("resendMissedMessages: %v", len(missedMessages))

	for _, mm := range missedMessages {

		log.Println("Player: ", playerID)
		log.Println("Player room: ", room)
		room.chBroadcast <- Broadcast{playersID: []string{playerID}, message: mm.message, msgNum: mm.msgNum}

		var dic struct {
			Turn string `json:"turn"`
		}
		// ako je turn message
		if err := json.Unmarshal(mm.message, &dic); err == nil {
			if dic.Turn == "rd" {
				// roll dice
				time.Sleep(1100 * time.Millisecond)
			} else {
				time.Sleep(200 * time.Millisecond)
			}
		}
	}
}

func otherPlayersID(playersID []string, outPlayerID string) []string {
	otherPlayersID := []string{}
	for _, playerID := range playersID {
		if playerID != outPlayerID {
			otherPlayersID = append(otherPlayersID, playerID)
		}
	}
	return otherPlayersID
}
