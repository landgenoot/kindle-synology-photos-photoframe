package main

import (
	"testing"
	"time"
)

func TestNextWakeupToday(t *testing.T) {
	now := time.Date(2009, time.November, 5, 5, 0, 0, 0, time.UTC)
	want := 3600
	got := nextWakeup(now, 6, 0)
	if got != want {
		t.Fatalf(`nextWakeup() = %v, want match for %#v, nil`, got, want)
	}
}

func TestNextWakeupTomorrow(t *testing.T) {
	now := time.Date(2009, time.November, 5, 7, 0, 0, 0, time.UTC)
	want := 82800
	got := nextWakeup(now, 6, 0)
	if got != want {
		t.Fatalf(`nextWakeup() = %v, want match for %#v, nil`, got, want)
	}
}

func TestParseBatteryLevel(t *testing.T) {
	want := 30
	got, err := parseBatteryLevel("30%")
	if got != want && err != nil {
		t.Fatalf(`parseBatteryLevel() = %v, %v, want match for %#v, nil`, got, err, want)
	}
}

func TestParseBadBatteryLevel(t *testing.T) {
	want := -1
	got, err := parseBatteryLevel("abc")
	if got != want && err != nil {
		t.Fatalf(`parseBatteryLevel() = %v, %v, want match for %#v, nil`, got, err, want)
	}
}
