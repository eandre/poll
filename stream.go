package main

import (
	"github.com/gorilla/websocket"
	"time"
)

var (
	TickInterval = 300 * time.Millisecond
)

type JSONEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func StreamVotes(ws *websocket.Conn, id uint64) {
	// Look up poll and use it
	poll, err := GetPoll(id)
	if err != nil {
		ws.WriteJSON(JSONEvent{"error", err.Error()})
		return
	}

	ticker := time.NewTicker(TickInterval)
	defer ticker.Stop()

	event := &JSONEvent{"data", nil}
	for _ = range ticker.C {
		event.Data = poll.Counts

		// Stop if we have an error or the poll is stopped
		if err := ws.WriteJSON(event); err != nil || poll.Stopped() {
			return
		}
	}
}
