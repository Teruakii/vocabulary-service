package word

import (
	"time"
)

type Word struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;column:word_id" json:"id"`
	Word        string    `gorm:"column:word;not null" json:"word"`
	Meaning     string    `gorm:"column:meaning;not null" json:"meaning"`
	OwnerUserID *uint64   `gorm:"column:owner_user_id" json:"ownerUserId"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Word) TableName() string {
	return "word"
}

func (w *Word) IsGlobal() bool { return w.OwnerUserID == nil }
