package userword

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
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, constants.ErrUserWordNotFound):
		response.ErrorMsg(c, http.StatusNotFound, "user word not found")
	case errors.Is(err, constants.ErrWordNotFound):
		response.ErrorMsg(c, http.StatusNotFound, "word not found")
	case errors.Is(err, constants.ErrUserWordDuplicate):
		response.ErrorMsg(c, http.StatusConflict, "word already added")
	case errors.Is(err, constants.ErrInvalidQuality):
		response.ErrorMsg(c, http.StatusBadRequest, "quality must be between 0 and 5")
	case errors.Is(err, constants.ErrWordEmpty), errors.Is(err, constants.ErrMeaningEmpty):
		response.Error(c, http.StatusBadRequest, err)
	case errors.Is(err, constants.ErrWordDuplicate):
		response.ErrorMsg(c, http.StatusConflict, "you already have a word with this text")
	default:
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
	}
}

func (h *Handler) AddWord(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AddWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	var (
		uw  *UserWord
		err error
	)
	switch {
	case req.WordID != nil:
		uw, err = h.svc.AddWord(c.Request.Context(), userID, *req.WordID)
	case req.Word != nil && req.Meaning != nil:
		uw, err = h.svc.AddCustomWord(c.Request.Context(), userID, *req.Word, *req.Meaning)
	default:
		response.ErrorMsg(c, http.StatusBadRequest, "provide either wordId, or word+meaning to create a new one")
		return
	}
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, uw)
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

	uw, err := h.svc.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, uw)
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	status := Status(c.Query("status"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	var categoryID *uint64
	if raw := c.Query("categoryId"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.ErrorMsg(c, http.StatusBadRequest, "invalid categoryId")
			return
		}
		categoryID = &id
	}

	items, hasMore, err := h.svc.List(c.Request.Context(), userID, status, categoryID, limit, offset)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, gin.H{"items": items, "hasMore": hasMore})
}

func (h *Handler) Due(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))

	items, err := h.svc.Due(c.Request.Context(), userID, limit)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, gin.H{"items": items})
}

func (h *Handler) Review(c *gin.Context) {
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

	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	uw, err := h.svc.Review(c.Request.Context(), userID, id, *req.Quality)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, uw)
}

func (h *Handler) ReviewByWord(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	wordID, err := strconv.ParseUint(c.Param("wordId"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid word id")
		return
	}

	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	uw, err := h.svc.ReviewByWord(c.Request.Context(), userID, wordID, *req.Quality)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, uw)
}

func (h *Handler) Remove(c *gin.Context) {
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

	if err := h.svc.Remove(c.Request.Context(), userID, id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.NoContent(c)
}
