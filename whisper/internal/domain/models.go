package domain

import "time"

type Cue struct {
	Start       time.Duration
	End         time.Duration
	Text        string
	Probability float64
	Tokens      []Token
	Origin      Origin
}

type Token struct {
	Start       time.Duration
	End         time.Duration
	Text        string
	Probability float64
	Origin      Origin
}

type Origin struct {
	Index int
	Start time.Duration
	End   time.Duration
}

type SpeechSegment struct {
	Start time.Duration
	End   time.Duration
}

type Chunk struct {
	Start time.Duration
	End   time.Duration
}
