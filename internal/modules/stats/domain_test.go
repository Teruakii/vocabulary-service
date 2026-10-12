package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecordActivity_StreakAcrossDays(t *testing.T) {
	day := func(d, h int) time.Time { return time.Date(2026, 10, d, h, 0, 0, 0, time.UTC) }
	s := &UserStats{}

	s.recordActivity(day(1, 9))
	assert.Equal(t, uint(1), s.CurrentStreak)

	s.recordActivity(day(1, 22)) // same day: unchanged
	assert.Equal(t, uint(1), s.CurrentStreak)

	s.recordActivity(day(2, 8))
	s.recordActivity(day(3, 23))
	assert.Equal(t, uint(3), s.CurrentStreak)
	// The longest streak must match the streak actually reached, not one more.
	assert.Equal(t, uint(3), s.LongestStreak)

	s.recordActivity(day(6, 10)) // gap: starts over
	assert.Equal(t, uint(1), s.CurrentStreak)
	assert.Equal(t, uint(3), s.LongestStreak)
	assert.Equal(t, day(6, 10), *s.LastActivityAt)
}

func TestRecordActivity_UsesCalendarDaysOfNowsLocation(t *testing.T) {
	bangkok := time.FixedZone("ICT", 7*60*60)
	// 23:00 UTC on the 1st is already the 2nd in Bangkok.
	last := time.Date(2026, 10, 1, 23, 0, 0, 0, time.UTC)
	s := &UserStats{CurrentStreak: 4, LongestStreak: 4, LastActivityAt: &last}

	s.recordActivity(time.Date(2026, 10, 2, 20, 0, 0, 0, bangkok))

	assert.Equal(t, uint(4), s.CurrentStreak)
}

func TestStreakAt_EndsAfterAMissedDay(t *testing.T) {
	last := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	s := &UserStats{CurrentStreak: 5, LastActivityAt: &last}

	assert.Equal(t, uint(5), s.streakAt(time.Date(2026, 10, 2, 23, 59, 0, 0, time.UTC)))
	assert.Equal(t, uint(0), s.streakAt(time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)))
}
