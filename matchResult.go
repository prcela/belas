package main

import (
	"time"
)

// MatchResult ...
type MatchResult struct {
	PlayersID     []string        `json:"players_id"`
	Scores        []int           `json:"scores"`
	WinnerID      string          `json:"winner_id"` // "drawn" or playerID
	WaitDurations []time.Duration `json:"wait_durations"`
	TotalWinnerID string          `json:"total_winner_id"` // winner after all (including wait time)
}

func newMatchResult(playersID []string, scores []int, waitDurations []time.Duration) *MatchResult {
	winnerID := "drawn"
	if scores[0] > scores[1] {
		winnerID = playersID[0]
	} else if scores[0] < scores[1] {
		winnerID = playersID[1]
	}
	totalWinnerID := winnerID
	if totalWinnerID == "drawn" {
		if waitDurations[0] > waitDurations[1] {
			totalWinnerID = playersID[0]
		} else {
			totalWinnerID = playersID[1]
		}
	}
	return &MatchResult{
		PlayersID:     playersID,
		Scores:        scores,
		WinnerID:      winnerID,
		WaitDurations: waitDurations,
		TotalWinnerID: totalWinnerID,
	}
}
