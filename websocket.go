package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"strconv"
	"time"
)

var (
	TickInterval = 300 * time.Millisecond
)

type JSONEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func StreamVotes(ws *websocket.Conn) {
	defer ws.Close()

	encoder := json.NewEncoder(ws)
	values := ws.Request().URL.Query()
	ids := values.Get("id")
	if ids == "" {
		encoder.Encode(JSONEvent{"error", "Could not parse id"})
		return
	}

	id, err := strconv.ParseUint(ids, 10, 64)
	if err != nil {
		encoder.Encode(JSONEvent{"error", err.Error()})
		return
	}

	// Look up poll and use it
	poll, err := GetPoll(id)
	if err != nil {
		encoder.Encode(JSONEvent{"error", err.Error()})
		return
	}

	ticker := time.NewTicker(TickInterval)
	defer ticker.Stop()

	event := &JSONEvent{"data", nil}
	for _ = range ticker.C {
		event.Data = poll.Counts

		// Stop if we have an error or the poll is stopped
		if err := encoder.Encode(event); err != nil || poll.Stopped() {
			return
		}
	}
}
