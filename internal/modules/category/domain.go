package category

import "time"

type Category struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:category_id" json:"id"`
	UserID    *uint64   `gorm:"column:user_id" json:"userId"`
	Name      string    `gorm:"column:category_name;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (Category) TableName() string { return "word_category" }

func (c *Category) IsGlobal() bool { return c.UserID == nil }

type Tag struct {
	VocabID    uint64 `gorm:"primaryKey;column:vocab_id" json:"vocabId"`
	CategoryID uint64 `gorm:"primaryKey;column:category_id" json:"categoryId"`
}

func (Tag) TableName() string { return "vocab_tags" }
