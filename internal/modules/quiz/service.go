package quiz

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/modules/userword"
)

const multipleChoiceOptionCount = 4

type StatsRecorder interface {
	RecordQuizCompleted(ctx context.Context, userID uint64) error
}

type service struct {
	repo  Repository
	words userword.Service
	stats StatsRecorder
}

func NewService(repo Repository, words userword.Service, stats StatsRecorder) Service {
	return &service{repo: repo, words: words, stats: stats}
}

func (s *service) Create(ctx context.Context, userID uint64, totalQuestions int, source Source) (*QuizDetail, error) {
	if totalQuestions <= 0 {
		return nil, constants.ErrInvalidQuizSize
	}
	if source == "" {
		source = SourceDue
	}
	if source != SourceDue && source != SourceAll {
		return nil, constants.ErrInvalidQuizSource
	}

	quizModel, questions, err := s.repo.CreateWithQuestions(ctx, userID, totalQuestions, source)
	if err != nil {
		if errors.Is(err, constants.ErrNoWordsAvailable) {
			return nil, constants.ErrNoWordsAvailable
		}
		return nil, fmt.Errorf("create quiz for user %d: %w", userID, err)
	}

	views := make([]QuestionView, 0, len(questions))
	for _, q := range questions {
		view, err := s.toPromptView(ctx, userID, q)
		if err != nil {
			return nil, fmt.Errorf("build question view: %w", err)
		}
		views = append(views, view)
	}

	return toQuizDetail(quizModel, views), nil
}

func (s *service) Get(ctx context.Context, userID, quizID uint64) (*QuizDetail, error) {
	quizModel, err := s.getOwnedQuiz(ctx, userID, quizID)
	if err != nil {
		return nil, err
	}

	questions, err := s.repo.ListQuestions(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("list questions for quiz %d: %w", quizID, err)
	}

	views := make([]QuestionView, 0, len(questions))
	for _, q := range questions {
		if q.Answered() {
			views = append(views, toAnsweredView(q))
			continue
		}
		view, err := s.toPromptView(ctx, userID, q)
		if err != nil {
			return nil, fmt.Errorf("build question view: %w", err)
		}
		views = append(views, view)
	}

	return toQuizDetail(quizModel, views), nil
}

func (s *service) List(ctx context.Context, userID uint64, limit, offset int) ([]QuizSummary, bool, error) {
	if limit <= 0 {
		limit = constants.DefaultListLimit
	}
	if offset < 0 {
		offset = 0
	}

	quizzes, hasMore, err := s.repo.ListQuizzesByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, false, fmt.Errorf("list quizzes for user %d: %w", userID, err)
	}

	summaries := make([]QuizSummary, 0, len(quizzes))
	for _, q := range quizzes {
		summaries = append(summaries, QuizSummary{
			ID:             q.ID,
			Score:          q.Score,
			TotalQuestions: q.TotalQuestions,
			CreatedAt:      q.CreatedAt,
			CompletedAt:    q.CompletedAt,
		})
	}
	return summaries, hasMore, nil
}

func (s *service) Answer(ctx context.Context, userID, quizID, questionID uint64, answer string) (*QuestionView, error) {
	quizModel, err := s.getOwnedQuiz(ctx, userID, quizID)
	if err != nil {
		return nil, err
	}
	if quizModel.IsCompleted() {
		return nil, constants.ErrQuizCompleted
	}

	q, err := s.repo.GetQuestion(ctx, questionID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrQuestionNotFound
		}
		return nil, fmt.Errorf("get question %d: %w", questionID, err)
	}
	if q.QuizID != quizID {
		return nil, constants.ErrQuestionNotFound
	}
	if q.Answered() {
		return nil, constants.ErrQuestionAlreadyAnswered
	}

	answer = strings.TrimSpace(answer)
	if answer == "" {
		return nil, constants.ErrAnswerEmpty
	}
	isCorrect := strings.EqualFold(answer, strings.TrimSpace(q.Meaning))

	if err := s.repo.SaveAnswer(ctx, questionID, answer, isCorrect); err != nil {
		return nil, fmt.Errorf("save answer for question %d: %w", questionID, err)
	}
	if isCorrect {
		if err := s.repo.IncrementScore(ctx, quizID); err != nil {
			return nil, fmt.Errorf("increment score for quiz %d: %w", quizID, err)
		}
	}

	if remaining, err := s.repo.CountUnanswered(ctx, quizID); err != nil {
		return nil, fmt.Errorf("count unanswered for quiz %d: %w", quizID, err)
	} else if remaining == 0 {
		if err := s.repo.CompleteQuiz(ctx, quizID); err != nil {
			return nil, fmt.Errorf("complete quiz %d: %w", quizID, err)
		}

		s.recordQuizCompleted(ctx, userID)
	}

	s.feedSRS(ctx, userID, q.WordID, isCorrect)

	q.UserAnswer = &answer
	q.IsCorrect = &isCorrect
	view := toAnsweredView(*q)
	return &view, nil
}

func (s *service) feedSRS(ctx context.Context, userID, wordID uint64, isCorrect bool) {
	quality := 1
	if isCorrect {
		quality = 4
	}
	_, _ = s.words.ReviewByWord(ctx, userID, wordID, quality)
}

func (s *service) recordQuizCompleted(ctx context.Context, userID uint64) {
	_ = s.stats.RecordQuizCompleted(ctx, userID)
}

func (s *service) Complete(ctx context.Context, userID, quizID uint64) (*QuizDetail, error) {
	quizModel, err := s.getOwnedQuiz(ctx, userID, quizID)
	if err != nil {
		return nil, err
	}
	if !quizModel.IsCompleted() {
		if err := s.repo.CompleteQuiz(ctx, quizID); err != nil {
			return nil, fmt.Errorf("complete quiz %d: %w", quizID, err)
		}
		s.recordQuizCompleted(ctx, userID)
	}
	return s.Get(ctx, userID, quizID)
}

func (s *service) getOwnedQuiz(ctx context.Context, userID, quizID uint64) (*Quiz, error) {
	q, err := s.repo.GetQuiz(ctx, quizID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrQuizNotFound
		}
		return nil, fmt.Errorf("get quiz %d: %w", quizID, err)
	}
	if q.UserID != userID {
		return nil, constants.ErrQuizNotFound
	}
	return q, nil
}

func (s *service) toPromptView(ctx context.Context, userID uint64, q questionWithWord) (QuestionView, error) {
	view := QuestionView{
		ID:           q.ID,
		WordID:       q.WordID,
		Word:         q.Word,
		QuestionType: q.QuestionType,
	}
	if q.QuestionType == QuestionMultipleChoice {
		distractors, err := s.repo.DistractorMeanings(ctx, userID, q.WordID, multipleChoiceOptionCount-1)
		if err != nil {
			return QuestionView{}, err
		}
		options := append(distractors, q.Meaning)
		rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
		view.Options = options
	}
	return view, nil
}

func toAnsweredView(q questionWithWord) QuestionView {
	correct := q.Meaning
	return QuestionView{
		ID:            q.ID,
		WordID:        q.WordID,
		Word:          q.Word,
		QuestionType:  q.QuestionType,
		Answered:      true,
		UserAnswer:    q.UserAnswer,
		CorrectAnswer: &correct,
		IsCorrect:     q.IsCorrect,
	}
}

func toQuizDetail(q *Quiz, views []QuestionView) *QuizDetail {
	return &QuizDetail{
		ID:             q.ID,
		Score:          q.Score,
		TotalQuestions: q.TotalQuestions,
		CreatedAt:      q.CreatedAt,
		CompletedAt:    q.CompletedAt,
		Questions:      views,
	}
}
