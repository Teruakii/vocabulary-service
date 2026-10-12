package stats

import "time"

type UserStats struct {
	UserID        uint64 `gorm:"primaryKey;column:user_id" json:"userId"`
	CurrentStreak uint   `gorm:"column:current_streak;not null;default:0" json:"currentStreak"`
	LongestStreak uint   `gorm:"column:longest_streak;not null;default:0" json:"longestStreak"`
	// TotalWordsLearned is how many of the user's words are mastered right now.
	// It is counted on read: a stored counter double counts a word that lapses
	// and is mastered again.
	TotalWordsLearned uint       `gorm:"-" json:"totalWordsLearned"`
	TotalQuizTaken    uint       `gorm:"column:total_quiz_taken;not null;default:0" json:"totalQuizTaken"`
	LastActivityAt    *time.Time `gorm:"column:last_activity_at" json:"lastActivityAt"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (UserStats) TableName() string {
	return "user_stats"
}

// recordActivity advances the streak for a learning action at now. Days are
// calendar days in now's location: a second action on the same day keeps the
// streak, one on the next day extends it, and a gap starts over at 1.
func (s *UserStats) recordActivity(now time.Time) {
	switch {
	case s.LastActivityAt != nil && sameDay(*s.LastActivityAt, now):
	case s.LastActivityAt != nil && sameDay(*s.LastActivityAt, now.AddDate(0, 0, -1)):
		s.CurrentStreak++
	default:
		s.CurrentStreak = 1
	}
	s.LongestStreak = max(s.LongestStreak, s.CurrentStreak)
	s.LastActivityAt = &now
}

// streakAt is the streak as of now: one whose last activity was before
// yesterday has ended, even though the stored value is only reset on the
// next activity.
func (s *UserStats) streakAt(now time.Time) uint {
	if s.LastActivityAt == nil ||
		sameDay(*s.LastActivityAt, now) || sameDay(*s.LastActivityAt, now.AddDate(0, 0, -1)) {
		return s.CurrentStreak
	}
	return 0
}

func sameDay(t, ref time.Time) bool {
	y1, m1, d1 := t.In(ref.Location()).Date()
	y2, m2, d2 := ref.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// DailyActivity counts how many learning actions (word reviews, including
// answers given in quizzes) a user did on one calendar day.
type DailyActivity struct {
	UserID       uint64    `gorm:"primaryKey;column:user_id"`
	ActivityDate time.Time `gorm:"primaryKey;column:activity_date;type:date"`
	Count        uint      `gorm:"column:count;not null;default:0"`
}

func (DailyActivity) TableName() string {
	return "user_activity"
}

// ActivityPoint is one day of the activity chart; Date is YYYY-MM-DD.
type ActivityPoint struct {
	Date  string `json:"date"`
	Count uint   `json:"count"`
}

const (
	DefaultActivityDays = 7
	MaxActivityDays     = 90
)
