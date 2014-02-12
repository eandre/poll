package main

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrUnknownPoll       = errors.New("Unknown poll id")
	ErrPollAlreadyExists = errors.New("Duplicate poll id")
	PollTimeout          = 1 * time.Hour
)

var (
	activePolls map[uint64]*Poll
	activeMutex sync.RWMutex
	activeCount uint64
)

func GetPoll(id uint64) (*Poll, error) {
	activeMutex.RLock()
	defer activeMutex.RUnlock()

	if poll, ok := activePolls[id]; ok {
		return poll, nil
	} else {
		return nil, ErrUnknownPoll
	}
}

func AddPoll(poll *Poll) uint64 {
	activeMutex.Lock()
	defer activeMutex.Unlock()

	activeCount++
	activePolls[activeCount] = poll
	go expirePoll(activeCount, poll)

	return activeCount
}

func expirePoll(id uint64, poll *Poll) {
	time.Sleep(PollTimeout)
	activeMutex.Lock()
	defer activeMutex.Unlock()

	poll.Stop()
	delete(activePolls, id)
}

func init() {
	activePolls = make(map[uint64]*Poll)
}
