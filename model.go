package main

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// StatItem ...
type StatItem struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	PlayerID  string        `json:"player_id" bson:"player_id"`
	Score     int           `json:"score" bson:"score"`
	Result    int           `json:"result" bson:"result"`
	Bet       int           `json:"bet" bson:"bet"`
	Timestamp time.Time     `json:"timestamp" bson:"timestamp"`
}

// PlayerScore ...
type PlayerScore struct {
	ID       string   `json:"id" bson:"_id,omitempty"`
	MaxScore int      `json:"max_score" bson:"maxScore"`
	AvgScore float64  `json:"avg_score" bson:"avgScore,omitempty"`
	Players  []Player `json:"players" bson:"players"` // nažalost, množina je ovdje jer se radi join sa lookupom
}

// ServerInfo ...
type ServerInfo struct {
	MinRequiredVersion int `json:"min_required_version"`
	RoomMainCt         int `json:"room_main_ct"`
	RoomMainFreeCt     int `json:"room_main_free_ct"`
}

// Action ...
type Action struct {
	name         string
	message      []byte
	fromPlayerID string
	msgNum       int32
}

// TurnTimeout ...
type TurnTimeout struct {
	winPlayerID string
	tableID     string
}

// Broadcast message to desired players
type Broadcast struct {
	playersID []string
	message   []byte
	msgNum    int32
}

type MissedMessage struct {
	message []byte
	msgNum  int32
}
