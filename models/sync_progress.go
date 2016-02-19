package models

import (
	"math"
)

type SyncProgress struct {
	Node       *Node
	Repository string
	State      string
	SizeTotal  int
	SizeLeft   int
	ItemsTotal int
	ItemsLeft  int
	Message    string
}

func (s *SyncProgress) ItemsDone() int {
	return s.ItemsTotal - s.ItemsLeft
}

func (s *SyncProgress) ItemsPercent() int {
	if s.ItemsTotal == 0 {
		return 100
	} else {
		return int(math.Floor(float64(s.ItemsDone()) / float64(s.ItemsTotal) * float64(100)))

	}
}

func (s *SyncProgress) SizeDone() int {
	return s.SizeTotal - s.SizeLeft
}

func (s *SyncProgress) SizePercent() int {
	if s.SizeTotal == 0 {
		return 100
	} else {
		return int(math.Floor(float64(s.SizeDone()) / float64(s.SizeTotal) * float64(100)))
	}
}
