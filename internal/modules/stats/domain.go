package stats

import "time"

type UserStats struct {
	UserID            uint64     `gorm:"primaryKey;column:user_id" json:"userId"`
	CurrentStreak     uint       `gorm:"column:current_streak;not null;default:0" json:"currentStreak"`
	LongestStreak     uint       `gorm:"column:longest_streak;not null;default:0" json:"longestStreak"`
	TotalWordsLearned uint       `gorm:"column:total_words_learned;not null;default:0" json:"totalWordsLearned"`
	TotalQuizTaken    uint       `gorm:"column:total_quiz_taken;not null;default:0" json:"totalQuizTaken"`
	LastActivityAt    *time.Time `gorm:"column:last_activity_at" json:"lastActivityAt"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (UserStats) TableName() string {
	return "user_stats"
}
