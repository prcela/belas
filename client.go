package main

import (
	"encoding/json"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline    = []byte{'\n'}
	space      = []byte{' '}
	msgCounter = int32(0)
)

func newMsgNum() int32 {
	msgCounter++
	return msgCounter
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	playerID string
	room     *Room
	conn     *websocket.Conn
	send     chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.room.chUnregisterClient <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.processMessage(message)
	}
}

func (c *Client) processMessage(message []byte) {

	var dic map[string]interface{}
	if err := json.Unmarshal(message, &dic); err != nil {
		panic(err)
	}
	log.Println(dic)

	if dic["msg_num"] == nil {
		dic["msg_num"] = float64(newMsgNum())
	} else {
		log.Println("msg_num is not nil")
	}

	color.Magenta("%v", dic)

	js, err := json.Marshal(dic)
	if err != nil {
		panic(err)
	}

	action := &Action{
		name:         dic["msg_func"].(string),
		message:      js,
		fromPlayerID: c.playerID,
		msgNum:       int32(dic["msg_num"].(float64)),
	}
	log.Println("Client action:", action.name)
	p := c.room.getPlayer(action.fromPlayerID)
	if p == nil {
		log.Println("client: from player is nil!")
		return
	}

	switch action.name {
	case "ack":
		c.room.chAction <- action

	case "ignore_invitation", "create_table", "join_table":
		if p.TableID == nil {
			c.room.chAction <- action
		}

	case "leave_table", "invite_player", "text_message", "turn", "end_match", "rematch":
		if p.TableID != nil {
			log.Println("c.room.action <- action")
			c.room.chAction <- action
		}
	}

	switch dic["msg_func"] {

	case "player_stat":

		var stat struct {
			LastN *int `json:"last_n,omitempty"`
		}
		if err := json.Unmarshal(message, &stat); err != nil {
			panic(err)
		}

		action := &Action{
			name:         "player_stat",
			message:      message,
			fromPlayerID: c.playerID,
		}
		c.room.chAction <- action

	case "stat_items":
		playerID := dic["player_id"].(string)
		log.Println("playerId: ", playerID)
		playerStatItems := c.room.statItems(playerID, nil)
		js, err := json.Marshal(struct {
			MsgFunc   string     `json:"msg_func"`
			PlayerID  string     `json:"player_id"`
			StatItems []StatItem `json:"stat_items"`
		}{
			MsgFunc:   "stat_items",
			PlayerID:  playerID,
			StatItems: playerStatItems,
		})

		if err != nil {
			log.Println(err)
		}

		c.send <- js

	case "update_player":
		c.updatePlayer(message)

	case "stat_item":
		c.addStatItem(message)

	case "room_info":
		c.send <- c.room.info()
	}

}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		log.Println("Client closed")
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(room *Room, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var cookie, _ = r.Cookie("playerId")

	client := &Client{room: room, conn: conn, playerID: cookie.Value, send: make(chan []byte, 256)}
	log.Println("Created client")

	client.room.chRegisterClient <- client
	log.Println("Registered client")

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (c *Client) updatePlayer(message []byte) {
	db, s := c.room.GetDatabaseSessionCopy()
	defer s.Close()

	var player *Player
	if err := json.Unmarshal(message, &player); err != nil {
		log.Println(err)
		return
	}
	change := bson.M{"$set": bson.M{
		"alias":      player.Alias,
		"diamonds":   player.Diamonds,
		"retentions": player.Retentions,
	}}

	db.C("players").Update(bson.M{"_id": player.ID}, change)
	c.room.chUpdatePlayer <- player
}

func (c *Client) addStatItem(message []byte) {
	db, s := c.room.GetDatabaseSessionCopy()
	defer s.Close()

	var statItem *StatItem
	if err := json.Unmarshal(message, &statItem); err != nil {
		log.Println(err)
		return
	}
	if err := db.C("statItems").Insert(statItem); err != nil {
		log.Println(err)
		return
	}
}
