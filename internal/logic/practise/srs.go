package practise

import (
	"math"
	"time"
)

// SRSResult holds the output of an SM-2 calculation.
type SRSResult struct {
	EaseFactor  float64
	Interval    int
	Repetitions int
	NextReview  time.Time
}

// SM2Calculate runs the SM-2 spaced repetition algorithm.
// quality: 0-5 (0=blank, 5=perfect)
func SM2Calculate(easeFactor float64, interval int, repetitions int, quality int) SRSResult {
	q := float64(quality)
	newEF := easeFactor + (0.1 - (5-q)*(0.08+(5-q)*0.02))
	if newEF < 1.3 {
		newEF = 1.3
	}

	var newInterval int
	var newRepetitions int

	if quality >= 3 {
		newRepetitions = repetitions + 1
		switch newRepetitions {
		case 1:
			newInterval = 1
		case 2:
			newInterval = 6
		default:
			newInterval = int(math.Ceil(float64(interval) * newEF))
		}
	} else {
		newRepetitions = 0
		newInterval = 1
	}

	if newInterval > 180 {
		newInterval = 180
	}

	nextReview := time.Now().Truncate(24 * time.Hour).
		Add(time.Duration(newInterval) * 24 * time.Hour)

	return SRSResult{
		EaseFactor:  math.Round(newEF*100) / 100,
		Interval:    newInterval,
		Repetitions: newRepetitions,
		NextReview:  nextReview,
	}
}

// ShouldFinish returns true when interval>=30 days and reps>=4 consecutive correct.
func ShouldFinish(interval int, repetitions int) bool {
	return interval >= 30 && repetitions >= 4
}
