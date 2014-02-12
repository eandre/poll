package main

import (
	"errors"
	"sync/atomic"
)

var (
	ErrInvalidAnswer   = errors.New("Invalid answer")
	ErrDuplicateAnswer = errors.New("Duplicate answer")
	ErrNoAnswers       = errors.New("No answers defined")
	ErrTooManyAnswers  = errors.New("Too many answers defined")
	ErrTooShort        = errors.New("Input too short")
	ErrTooLong         = errors.New("Input too long")
)

const (
	MaxAnswers = 9
	MinLength  = 1
	MaxLength  = 127
)

type Poll struct {
	Question string   `json:"question"`
	Answers  []string `json:"answers"`
	Counts   []uint32 `json:"counts"`
	stopped  int32
}

func checkLength(s string) error {
	if len(s) < MinLength {
		return ErrTooShort
	} else if len(s) > MaxLength {
		return ErrTooLong
	}
	return nil
}

func NewPoll(question string, answers ...string) (*Poll, error) {
	if len(answers) == 0 {
		return nil, ErrNoAnswers
	} else if len(answers) > MaxAnswers {
		return nil, ErrTooManyAnswers
	} else if err := checkLength(question); err != nil {
		return nil, err
	}

	poll := &Poll{
		Question: question,
		Answers:  make([]string, len(answers)),
		Counts:   make([]uint32, len(answers)),
		stopped:  0,
	}

	for i, answer := range answers {
		if err := checkLength(answer); err != nil {
			return nil, err
		}
		poll.Answers[i] = answer
	}

	return poll, nil
}

func (p *Poll) RecordAnswers(indices ...int) error {
	numIndices := len(p.Answers)
	for i, idx := range indices {
		// Make sure the index is valid
		if idx >= numIndices {
			return ErrInvalidAnswer
		}

		// Make sure this index has not been seen before
		for j := 0; j < i; j++ {
			if indices[j] == idx {
				return ErrDuplicateAnswer
			}
		}

		// Increment the count
		atomic.AddUint32(&p.Counts[idx], 1)
	}

	return nil
}

func (p *Poll) Stop() {
	atomic.StoreInt32(&p.stopped, 1)
}

func (p *Poll) Stopped() bool {
	return p.stopped == 1
}
