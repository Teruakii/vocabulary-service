package quiz

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/infra/http/response"
	"vocabulary-service/internal/middleware"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	detail, err := h.service.Create(c.Request.Context(), userID, req.TotalQuestions, Source(req.Source))
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrInvalidQuizSize), errors.Is(err, constants.ErrInvalidQuizSource):
			response.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, constants.ErrNoWordsAvailable):
			response.ErrorMsg(c, http.StatusConflict, "no words available to quiz on")
		default:
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	response.OK(c, http.StatusCreated, detail)
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, hasMore, err := h.service.List(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"items": items, "hasMore": hasMore})
}

func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid id")
		return
	}

	detail, err := h.service.Get(c.Request.Context(), userID, id)
	if err != nil {
		if errors.Is(err, constants.ErrQuizNotFound) {
			response.ErrorMsg(c, http.StatusNotFound, "not found")
			return
		}
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, detail)
}

func (h *Handler) Answer(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	quizID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid quiz id")
		return
	}
	questionID, err := strconv.ParseUint(c.Param("qid"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid question id")
		return
	}

	var req AnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	view, err := h.service.Answer(c.Request.Context(), userID, quizID, questionID, req.Answer)
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrQuizNotFound), errors.Is(err, constants.ErrQuestionNotFound):
			response.ErrorMsg(c, http.StatusNotFound, "not found")
		case errors.Is(err, constants.ErrQuizCompleted):
			response.ErrorMsg(c, http.StatusConflict, "quiz already completed")
		case errors.Is(err, constants.ErrQuestionAlreadyAnswered):
			response.ErrorMsg(c, http.StatusConflict, "question already answered")
		case errors.Is(err, constants.ErrAnswerEmpty):
			response.Error(c, http.StatusBadRequest, err)
		default:
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	response.OK(c, http.StatusOK, view)
}

func (h *Handler) Complete(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid id")
		return
	}

	detail, err := h.service.Complete(c.Request.Context(), userID, id)
	if err != nil {
		if errors.Is(err, constants.ErrQuizNotFound) {
			response.ErrorMsg(c, http.StatusNotFound, "not found")
			return
		}
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, detail)
}
