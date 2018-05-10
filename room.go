// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"github.com/fatih/color"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"sync"
	"time"
)

// Room maintains the set of active clients and broadcasts messages to the clients.
type Room struct {
	name string
	mu   sync.Mutex

	// channels
	chBroadcastAll       chan []byte
	chBroadcast          chan Broadcast
	chVerifyBroadcastNum chan int32
	chAction             chan *Action

	chRegisterClient   chan *Client
	chUnregisterClient chan *Client
	chWaitClient       chan *Client

	chRegisterPlayer   chan *Player
	chUnregisterPlayer chan *Player
	chUpdatePlayer     chan *Player

	chTurnTimeout chan TurnTimeout

	clients         map[*Client]bool
	players         map[string]*Player
	tables          map[string]*Table
	broadcastsToAck map[int32]*Broadcast
}

func newRoom(name string) *Room {
	return &Room{
		name:                 name,
		chBroadcastAll:       make(chan []byte, 256),
		chBroadcast:          make(chan Broadcast, 256),
		chVerifyBroadcastNum: make(chan int32),
		chRegisterClient:     make(chan *Client),
		chUnregisterClient:   make(chan *Client),
		chWaitClient:         make(chan *Client),
		chRegisterPlayer:     make(chan *Player),
		chUnregisterPlayer:   make(chan *Player),
		chUpdatePlayer:       make(chan *Player),
		chTurnTimeout:        make(chan TurnTimeout),
		chAction:             make(chan *Action),
		clients:              make(map[*Client]bool),
		players:              make(map[string]*Player),
		tables:               make(map[string]*Table),
		broadcastsToAck:      make(map[int32]*Broadcast),
	}
}

// GetDatabaseSessionCopy create a copy of opened session
func (room *Room) GetDatabaseSessionCopy() (*mgo.Database, *mgo.Session) {
	sessionCopy := session.Copy()
	return sessionCopy.DB(room.name), sessionCopy
}

func (room *Room) run() {
	log.Println("Room run ", room.name)

	for {
		select {
		case client := <-room.chRegisterClient:
			room.mu.Lock()
			room.clients[client] = true
			room.mu.Unlock()

		case client := <-room.chUnregisterClient:
			log.Println("room.unregisterClient")
			room.mu.Lock()
			if _, ok := room.clients[client]; ok {
				// remove player if it is not for table
				if client.playerID != nil {
					if player := room.players[*client.playerID]; player != nil {
						log.Printf("unregister client: player: %v, tableID: %v", player.Alias, player.TableID)
						if !player.waitingForClient {
							if player.TableID != nil {
								go player.waitForClientConnection(room, 30*time.Second)
							} else {
								go player.waitForClientConnection(room, 5*time.Second)
							}
						}
					}
				}
				delete(room.clients, client)
				close(client.send)
			}
			room.mu.Unlock()

		case client := <-room.chWaitClient:

			if player := room.players[*client.playerID]; player != nil {
				go func() {
					if player.waitingForClient {
						player.chWaitClient <- client
					}
					room.resendMissedMessages(player.ID, player.missedMessages)
				}()
				room.mu.Lock()
				player.missedMessages = []MissedMessage{}
				room.mu.Unlock()
			}

		case player := <-room.chRegisterPlayer:
			log.Println("Register player:", player)
			room.mu.Lock()
			room.players[player.ID] = player
			room.mu.Unlock()
			room.chBroadcastAll <- room.info()

		case player := <-room.chUnregisterPlayer:
			log.Printf("Unregister player: %v table: %v\n", player.ID, player.TableID)

			if player.TableID != nil {
				if table := room.tables[*player.TableID]; table != nil {
					if m := table.Match; m != nil {

						m.leave(player.ID)
						table.unregisterMatch()

						msgNum := newMsgNum()
						js, err := json.Marshal(struct {
							MsgFunc string `json:"msg_func"`
							MatchID string `json:"match_id"`
							TableID string `json:"table_id"`
							MsgNum  int32  `json:"msg_num"`
						}{
							MsgFunc: "opponent_disconnected",
							MatchID: m.ID,
							TableID: table.ID,
							MsgNum:  msgNum,
						})
						if err != nil {
							log.Println(err)
						}

						otherPlayersID := m.PlayersID
						log.Println("Other players:", otherPlayersID)

						room.chBroadcast <- Broadcast{playersID: otherPlayersID, message: js, msgNum: msgNum}
					}
				}
			}

			room.mu.Lock()
			room.cleanBroadcastsToAck(player.ID)
			delete(room.players, player.ID)
			room.mu.Unlock()

			room.chBroadcastAll <- room.info()

		case player := <-room.chUpdatePlayer:
			room.mu.Lock()
			if up := room.players[player.ID]; up != nil {
				up.Alias = player.Alias
				up.Diamonds = player.Diamonds
				up.Retentions = player.Retentions
			}
			room.mu.Unlock()

		case message := <-room.chBroadcastAll:
			for client := range room.clients {
				select {
				case client.send <- message:
				default:
					go func(c *Client) {
						room.chUnregisterClient <- client
					}(client)
				}
			}

		case msgNum := <-room.chVerifyBroadcastNum:
			if b := room.broadcastsToAck[msgNum]; b != nil {
				// try broadcast again to left players
				color.HiGreen("Resending broadcast %v", msgNum)
				room.chBroadcast <- *b
			} else {
				color.Green("Verify broadcast %v ✓ Done", msgNum)
			}

		case broadcast := <-room.chBroadcast:
			if broadcast.msgNum != 0 {
				room.broadcastsToAck[broadcast.msgNum] = &broadcast
				go func(msgNum int32) {
					time.Sleep(5 * time.Second)
					room.chVerifyBroadcastNum <- msgNum
				}(broadcast.msgNum)
			}
			for _, playerID := range broadcast.playersID {
				if player := room.players[playerID]; player != nil {
					foundClient := false
					for client := range room.clients {
						if client.playerID != nil {
							if playerID == *client.playerID {
								foundClient = true
								select {
								case client.send <- broadcast.message:
								default:
									if broadcast.msgNum != 0 {
										color.Yellow("append missed message")
										player.missedMessages = append(player.missedMessages, MissedMessage{message: broadcast.message, msgNum: broadcast.msgNum})
									}
									go func(c *Client) {
										room.chUnregisterClient <- client
									}(client)
								}
							}
						}
					}
					if broadcast.msgNum != 0 {
						if foundClient {
							room.mu.Lock()
							player.toAck[broadcast.msgNum] = true
							room.mu.Unlock()
							color.Yellow("player: %v, to ack: %v", player.Alias, player.toAck)
						} else {
							log.Println("Client not found")
							color.Yellow("append missed message")
							player.missedMessages = append(player.missedMessages, MissedMessage{message: broadcast.message, msgNum: broadcast.msgNum})
						}
					}
				}
			}

		case turnTimeout := <-room.chTurnTimeout:
			var otherPlayerID string
			if player := room.players[turnTimeout.winPlayerID]; player != nil && player.TableID != nil && *player.TableID == turnTimeout.tableID {
				if table := room.tables[*player.TableID]; table != nil && table.Match != nil {
					for _, playerID := range table.PlayersID {
						if playerID != turnTimeout.winPlayerID {
							otherPlayerID = playerID
						}
					}

					if m := table.Match; m != nil {
						m.leave(otherPlayerID)
						table.unregisterMatch()
					}

					msgNum := newMsgNum()
					js, err := json.Marshal(struct {
						MsgFunc     string `json:"msg_func"`
						Table       *Table `json:"table"`
						WinPlayerID string `json:"win_player_id"`
						MsgNum      int32  `json:"msg_num"`
					}{
						MsgFunc:     "turn_timeout",
						Table:       table,
						WinPlayerID: turnTimeout.winPlayerID,
						MsgNum:      msgNum,
					})
					if err != nil {
						log.Println(err)
					}

					table.room.chBroadcast <- Broadcast{playersID: table.PlayersID, message: js, msgNum: msgNum}

					table.leave(otherPlayerID)
				}
			}

		case action := <-room.chAction:

			switch action.name {

			case "ack":
				room.ackAction(action)

			case "player_stat":
				room.playerStatAction(action)

			case "create_table":
				room.createTableAction(action)

			case "join_table", "rematch":
				room.forwardToTableAction(action)

				room.chBroadcastAll <- room.info()

			case "end_match", "leave_table":
				room.forwardToTableAction(action)

				room.chBroadcastAll <- room.info()

			case "turn":
				log.Println("room case turn")
				room.forwardToTableAction(action)

			case "invite_player":
				room.invitePlayerAction(action)

			case "ignore_invitation":
				room.ignoreInvitationAction(action)

			case "text_message":
				room.textMessageAction(action)
			}
		}
	}
}

func (room *Room) getPlayer(id string) *Player {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.players[id]
}

func (room *Room) info() []byte {
	room.mu.Lock()
	defer room.mu.Unlock()
	js, err := json.Marshal(struct {
		MsgFunc string             `json:"msg_func"`
		Players map[string]*Player `json:"players"`
		Tables  map[string]*Table  `json:"tables"`
	}{
		MsgFunc: "room_info",
		Players: room.players,
		Tables:  room.tables,
	})
	if err != nil {
		log.Println(err)
	}
	return js
}

func (room *Room) ackAction(action *Action) {
	room.mu.Lock()
	if p := room.players[action.fromPlayerID]; p != nil {
		delete(p.toAck, action.msgNum)
		color.Yellow("player: %v, to ack: %v", p.Alias, p.toAck)
	}
	if b := room.broadcastsToAck[action.msgNum]; b != nil {
		otherPlayersID := []string{}
		for _, playerID := range b.playersID {
			if playerID != action.fromPlayerID {
				otherPlayersID = append(otherPlayersID, playerID)
			}
		}
		b.playersID = otherPlayersID
		if len(b.playersID) == 0 {
			color.Blue("deleted broadcast %v", b.msgNum)
			delete(room.broadcastsToAck, b.msgNum)
		}
	}
	room.mu.Unlock()
}

func (room *Room) playerStatAction(action *Action) {

	db, s := room.GetDatabaseSessionCopy()
	defer s.Close()

	playerID := action.fromPlayerID

	log.Println("playerId: ", playerID)

	player := room.getPlayer(playerID)

	foundInRoom := player != nil

	if foundInRoom {
		log.Println("found player in room", playerID)
		log.Println("wait channel: ", player.chWaitClient)
		// if still waiting for client
		for c := range room.clients {
			if c.playerID != nil && *c.playerID == playerID {
				go func(c *Client) {
					color.Cyan("room.waitClient <- c")
					room.chWaitClient <- c
				}(c)
				break
			}
		}

	} else {
		log.Println("Not found player in room")
		err := db.C("players").FindId(playerID).One(&player)
		if err == nil {
			player.room = room
			player.missedMessages = []MissedMessage{}
			player.chWaitClient = make(chan *Client)
			player.toAck = make(map[int32]bool)
		} else {
			log.Printf("Not found player %v in database\n", playerID)
			log.Println(err)
			log.Println("New player:", playerID)
			player = newPlayer(room, playerID)
			db.C("players").Insert(player)
		}

		room.mu.Lock()
		room.players[player.ID] = player
		room.mu.Unlock()
	}

	var stat struct {
		LastN *int `json:"last_n,omitempty"`
	}
	if err := json.Unmarshal(action.message, &stat); err != nil {
		panic(err)
	}
	playerStatItems := room.statItems(playerID, stat.LastN)

	js, err := json.Marshal(struct {
		MsgFunc     string     `json:"msg_func"`
		Player      *Player    `json:"player"`
		StatItems   []StatItem `json:"stat_items"`
		FoundInRoom bool       `json:"found_in_room"`
	}{
		MsgFunc:     "player_stat",
		Player:      player,
		StatItems:   playerStatItems,
		FoundInRoom: foundInRoom,
	})

	if err != nil {
		log.Println(err)
	}

	room.chBroadcast <- Broadcast{
		playersID: []string{playerID},
		message:   js,
	}

	if !foundInRoom {
		room.chBroadcastAll <- room.info()
	}

}

func (room *Room) createTableAction(action *Action) {
	p := room.getPlayer(action.fromPlayerID)
	if p == nil || p.TableID != nil {
		return
	}

	var dic struct {
		TurnDuration int      `json:"turn_duration"`
		Bet          int64    `json:"bet"`
		Private      bool     `json:"private"`
		PlayersID    []string `json:"players_id"`
		GameType     int      `json:"game_type"`
		UpToPoints   int      `json:"up_to_points"`
	}
	if err := json.Unmarshal(action.message, &dic); err != nil {
		panic(err)
	}
	table := newTable(room, 4, dic.PlayersID, dic.Bet, dic.Private, dic.GameType, dic.UpToPoints)
	room.mu.Lock()

	p.TableID = &table.ID
	room.tables[table.ID] = table
	room.mu.Unlock()

	room.chBroadcastAll <- room.info()

}

func (room *Room) forwardToTableAction(action *Action) {

	var dic struct {
		TableID string `json:"table_id"`
	}
	if err := json.Unmarshal(action.message, &dic); err != nil {
		panic(err)
	}
	log.Println("Forward to tableID:", dic.TableID)
	log.Println("U table stigo action:", action.name)

	if table := room.tables[dic.TableID]; table != nil {
		switch action.name {
		case "join_table":
			table.joinAction(action)

		case "leave_table":
			table.leaveAction(action)
			if len(table.PlayersID) == 0 {
				room.unregisterTable(table)
			}
		case "end_match", "turn":
			log.Println("table.Match.chAction <- action  match:", table.Match)
			table.forwardToMatchAction(action)
		case "rematch":
			table.rematchAction(action)
		}
	} else {
		log.Println("Table is nil!!!!")
		log.Println(room.tables)
	}
}

func (room *Room) unregisterTable(table *Table) {
	log.Println("Unregister table:", table.ID)
	room.mu.Lock()
	for _, playerID := range table.PlayersID {
		if p := room.players[playerID]; p != nil && p.TableID != nil {
			if *p.TableID == table.ID {
				p.TableID = nil
			}
		}
	}
	delete(room.tables, table.ID)
	room.mu.Unlock()

	room.chBroadcastAll <- room.info()
}

func (room *Room) invitePlayerAction(action *Action) {
	var dic struct {
		Sender    string `json:"sender"`
		Recipient string `json:"recipient"`
	}

	if err := json.Unmarshal(action.message, &dic); err != nil {
		panic(err)
	}

	room.chBroadcast <- Broadcast{playersID: []string{dic.Recipient}, message: action.message}

}

func (room *Room) ignoreInvitationAction(action *Action) {
	var dic struct {
		Sender    string `json:"sender"`
		Recipient string `json:"recipient"`
	}

	if err := json.Unmarshal(action.message, &dic); err != nil {
		panic(err)
	}

	room.chBroadcast <- Broadcast{playersID: []string{dic.Sender}, message: action.message}

}

func (room *Room) textMessageAction(action *Action) {
	var dic struct {
		Sender    string `json:"sender"`
		Recipient string `json:"recipient"`
		Text      string `json:"text"`
	}

	if err := json.Unmarshal(action.message, &dic); err != nil {
		panic(err)
	}

	room.chBroadcast <- Broadcast{playersID: []string{dic.Recipient}, message: action.message}
}

func (room *Room) cleanBroadcastsToAck(playerID string) {
	for _, b := range room.broadcastsToAck {
		b.playersID = otherPlayersID(b.playersID, playerID)
		if len(b.playersID) == 0 {
			color.Blue("deleted broadcast %v", b.msgNum)
			delete(room.broadcastsToAck, b.msgNum)
		}
	}
}

func (room *Room) statItems(playerID string, lastN *int) []StatItem {
	db, s := room.GetDatabaseSessionCopy()
	defer s.Close()

	statItems := []StatItem{}
	pipeline := []bson.M{
		{"$match": bson.M{"player_id": playerID}},
	}
	if lastN != nil {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"timestamp": -1}})
		pipeline = append(pipeline, bson.M{"$limit": *lastN})
	}
	err := db.C("statItems").Pipe(pipeline).All(&statItems)
	if err != nil {
		log.Println(err)
	}
	return statItems
}
