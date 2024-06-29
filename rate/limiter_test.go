package rate

import (
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	burst := 1

	interval := 10 * time.Millisecond
	lim := Every(interval)
	r := NewLimiter(burst, 100, lim)

	tooshort := 1 * time.Millisecond

	client := "test@test.com"
	expected := []bool{true, false, true, true, false, false}
	waits := []time.Duration{tooshort, interval, interval, tooshort, tooshort, tooshort}
	for i, exp := range expected {
		if got := r.Check(client); got != exp {
			t.Fatalf("iteration %d: expected %v, but got %v", i, exp, got)
		}
		time.Sleep(waits[i])
	}
}

func TestLimiterWithBurst(t *testing.T) {
	client := "test@test.com"
	burst := 10

	interval := 100 * time.Millisecond
	lim := Every(interval)

	tooshort := 10 * time.Millisecond

	shortest := 1 * time.Millisecond

	expected := []bool{true, true, true, true, true, true, true, true, true, true}
	waits := []time.Duration{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	expected = append(expected, false, true, true, false, false, false)
	waits = append(waits, interval, interval, tooshort, tooshort, shortest, shortest)

	rr := NewLimiter(burst, 100, lim)
	for i, exp := range expected {
		if got := rr.Check(client); got != exp {
			t.Fatalf("iteration %d: expected %v, but got %v", i, exp, got)
		}
		time.Sleep(waits[i])
	}
}
