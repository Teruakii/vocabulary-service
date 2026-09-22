package quiz

import "time"

type CreateQuizRequest struct {
	TotalQuestions int    `json:"totalQuestions" binding:"required,min=1,max=50"`
	Source         string `json:"source"`
}

type AnswerRequest struct {
	Answer string `json:"answer" binding:"required"`
}

type QuestionView struct {
	ID            uint64       `json:"id"`
	WordID        uint64       `json:"wordId"`
	Word          string       `json:"word"`
	QuestionType  QuestionType `json:"questionType"`
	Options       []string     `json:"options,omitempty"`
	Answered      bool         `json:"answered"`
	UserAnswer    *string      `json:"userAnswer,omitempty"`
	CorrectAnswer *string      `json:"correctAnswer,omitempty"`
	IsCorrect     *bool        `json:"isCorrect,omitempty"`
}

type QuizDetail struct {
	ID             uint64         `json:"id"`
	Score          uint           `json:"score"`
	TotalQuestions uint           `json:"totalQuestions"`
	CreatedAt      time.Time      `json:"createdAt"`
	CompletedAt    *time.Time     `json:"completedAt"`
	Questions      []QuestionView `json:"questions"`
}

type QuizSummary struct {
	ID             uint64     `json:"id"`
	Score          uint       `json:"score"`
	TotalQuestions uint       `json:"totalQuestions"`
	CreatedAt      time.Time  `json:"createdAt"`
	CompletedAt    *time.Time `json:"completedAt"`
}
