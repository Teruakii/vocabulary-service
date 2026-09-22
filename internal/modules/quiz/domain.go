package quiz

import "time"

type Quiz struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:quiz_id" json:"id"`
	UserID         uint64     `gorm:"column:user_id;not null" json:"userId"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	Score          uint       `gorm:"column:score;not null;default:0" json:"score"`
	TotalQuestions uint       `gorm:"column:total_questions;not null" json:"totalQuestions"`
	CompletedAt    *time.Time `gorm:"column:completed_at" json:"completedAt"`
}

func (Quiz) TableName() string {
	return "quizzes"
}

func (q *Quiz) IsCompleted() bool {
	return q.CompletedAt != nil
}

type QuestionType string

const (
	QuestionTyped          QuestionType = "typed"
	QuestionMultipleChoice QuestionType = "multiple_choice"
)

type Source string

const (
	SourceDue Source = "due"
	SourceAll Source = "all"
)

type Question struct {
	ID           uint64       `gorm:"primaryKey;autoIncrement;column:quiz_question_id" json:"id"`
	QuizID       uint64       `gorm:"column:quiz_id;not null" json:"quizId"`
	WordID       uint64       `gorm:"column:word_id;not null" json:"wordId"`
	QuestionType QuestionType `gorm:"column:question_type;not null;default:typed" json:"questionType"`
	UserAnswer   *string      `gorm:"column:user_answer" json:"-"`
	IsCorrect    *bool        `gorm:"column:is_correct" json:"-"`
}

func (Question) TableName() string {
	return "quiz_question"
}

func (q *Question) Answered() bool {
	return q.IsCorrect != nil
}

type questionWithWord struct {
	Question
	Word    string `gorm:"column:word"`
	Meaning string `gorm:"column:meaning"`
}
