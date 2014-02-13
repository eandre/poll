package main

import (
	"errors"
	"github.com/willf/bloom"
	"sync"
	"sync/atomic"
)

var (
	ErrInvalidAnswer   = errors.New("Invalid answer")
	ErrDuplicateAnswer = errors.New("Duplicate answer")
	ErrNoAnswers       = errors.New("No answers defined")
	ErrTooManyAnswers  = errors.New("Too many answers defined")
	ErrTooShort        = errors.New("Input too short")
	ErrTooLong         = errors.New("Input too long")
	ErrAlreadyVoted    = errors.New("You have already voted")
)

const (
	MaxAnswers = 9
	MinLength  = 1
	MaxLength  = 127
)

// Bloom filter definition; about p=0.0001, n=1000
const (
	FilterM = 19171
	FilterK = 13
)

type Poll struct {
	MultipleChoice bool     `json:"multipleChoice"`
	Question       string   `json:"question"`
	Answers        []string `json:"answers"`
	Counts         []uint32 `json:"counts"`
	stopped        int32
	filter         *bloom.BloomFilter
	filterMutex    sync.Mutex
}

func checkLength(s string) error {
	if len(s) < MinLength {
		return ErrTooShort
	} else if len(s) > MaxLength {
		return ErrTooLong
	}
	return nil
}

func NewPoll(checkDuplicates, multipleChoice bool, question string, answers ...string) (*Poll, error) {
	if len(answers) == 0 {
		return nil, ErrNoAnswers
	} else if len(answers) > MaxAnswers {
		return nil, ErrTooManyAnswers
	} else if err := checkLength(question); err != nil {
		return nil, err
	}

	poll := &Poll{
		MultipleChoice: multipleChoice,
		Question:       question,
		Answers:        make([]string, len(answers)),
		Counts:         make([]uint32, len(answers)),
		stopped:        0,
	}

	if checkDuplicates {
		poll.filter = bloom.New(FilterM, FilterK)
	}

	for i, answer := range answers {
		if err := checkLength(answer); err != nil {
			return nil, err
		}
		poll.Answers[i] = answer
	}

	return poll, nil
}

func (p *Poll) RecordOrigin(key []byte) bool {
	if p.filter == nil {
		return true
	}

	p.filterMutex.Lock()
	defer p.filterMutex.Unlock()

	success := !p.filter.Test(key)
	p.filter.Add(key)
	return success
}

func (p *Poll) RecordAnswers(indices ...uint32) error {
	if len(indices) == 0 {
		return ErrNoAnswers
	} else if len(indices) != 1 && !p.MultipleChoice {
		return ErrTooManyAnswers
	}

	numIndices := uint32(len(p.Answers))

	// Validate
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
	}

	// Increment the counts
	for _, idx := range indices {
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
