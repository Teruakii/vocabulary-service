package userword

import "time"

type Status string

const (
	StatusNew      Status = "new"
	StatusLearning Status = "learning"
	StatusMastered Status = "mastered"
)

type UserWord struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:user_word_id" json:"id"`
	UserID         uint64     `gorm:"column:user_id;not null" json:"userId"`
	WordID         uint64     `gorm:"column:word_id;not null" json:"wordId"`
	Status         Status     `gorm:"column:status;not null;default:new" json:"status"`
	EaseFactor     float64    `gorm:"column:ease_factor;not null;default:2.50" json:"easeFactor"`
	IntervalDays   uint       `gorm:"column:interval_days;not null;default:0" json:"intervalDays"`
	Repetitions    uint       `gorm:"column:repetitions;not null;default:0" json:"repetitions"`
	NextReviewAt   *time.Time `gorm:"column:next_review_at" json:"nextReviewAt"`
	LastReviewedAt *time.Time `gorm:"column:last_reviewed_at" json:"lastReviewedAt"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"createdAt"`
}

func (UserWord) TableName() string {
	return "user_word"
}

type UserWordDetail struct {
	UserWord
	Word    string `gorm:"column:word" json:"word"`
	Meaning string `gorm:"column:meaning" json:"meaning"`
}
