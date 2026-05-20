package practise

import (
	"testing"
	"time"
)

func TestSM2Calculate_FirstCorrect(t *testing.T) {
	r := SM2Calculate(2.5, 0, 0, 5)
	if r.Repetitions != 1 {
		t.Errorf("expected repetitions=1, got %d", r.Repetitions)
	}
	if r.Interval != 1 {
		t.Errorf("expected interval=1, got %d", r.Interval)
	}
	if r.EaseFactor < 2.5 {
		t.Errorf("expected EF >= 2.5, got %f", r.EaseFactor)
	}
}

func TestSM2Calculate_SecondCorrect(t *testing.T) {
	r := SM2Calculate(2.6, 1, 1, 5)
	if r.Repetitions != 2 {
		t.Errorf("expected repetitions=2, got %d", r.Repetitions)
	}
	if r.Interval != 6 {
		t.Errorf("expected interval=6, got %d", r.Interval)
	}
}

func TestSM2Calculate_ThirdCorrect(t *testing.T) {
	r := SM2Calculate(2.7, 6, 2, 5)
	if r.Repetitions != 3 {
		t.Errorf("expected repetitions=3, got %d", r.Repetitions)
	}
	// ceil(6 * 2.8) = 17
	if r.Interval != 17 {
		t.Errorf("expected interval=17, got %d", r.Interval)
	}
}

func TestSM2Calculate_Wrong(t *testing.T) {
	r := SM2Calculate(2.5, 6, 3, 2)
	if r.Repetitions != 0 {
		t.Errorf("expected repetitions=0, got %d", r.Repetitions)
	}
	if r.Interval != 1 {
		t.Errorf("expected interval=1, got %d", r.Interval)
	}
}

func TestSM2Calculate_EFFloor(t *testing.T) {
	r := SM2Calculate(1.3, 1, 1, 0)
	if r.EaseFactor < 1.3 {
		t.Errorf("EF should not drop below 1.3, got %f", r.EaseFactor)
	}
}

func TestSM2Calculate_IntervalCap(t *testing.T) {
	r := SM2Calculate(2.8, 100, 5, 5)
	if r.Interval > 180 {
		t.Errorf("interval should cap at 180, got %d", r.Interval)
	}
}

func TestSM2Calculate_NextReviewInFuture(t *testing.T) {
	r := SM2Calculate(2.5, 0, 0, 4)
	if r.NextReview.Before(time.Now()) {
		t.Errorf("next review should be in the future")
	}
}

func TestShouldFinish_True(t *testing.T) {
	if !ShouldFinish(30, 4) {
		t.Error("interval>=30 and reps>=4 should finish")
	}
}

func TestShouldFinish_False_LowInterval(t *testing.T) {
	if ShouldFinish(6, 4) {
		t.Error("interval<30 should not finish")
	}
}

func TestShouldFinish_False_LowReps(t *testing.T) {
	if ShouldFinish(30, 3) {
		t.Error("reps<4 should not finish")
	}
}
