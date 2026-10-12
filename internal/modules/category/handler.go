package category

import (
	"errors"
	"fmt"
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
	case errors.Is(err, constants.ErrCategoryNotFound):
		response.ErrorMsg(c, http.StatusNotFound, "category not found")
	case errors.Is(err, constants.ErrVocabNotFound):
		response.ErrorMsg(c, http.StatusNotFound, "vocab entry not found")
	case errors.Is(err, constants.ErrCategoryForbidden):
		response.ErrorMsg(c, http.StatusForbidden, "cannot modify this category")
	case errors.Is(err, constants.ErrTagDuplicate):
		response.ErrorMsg(c, http.StatusConflict, "word already tagged with this category")
	case errors.Is(err, constants.ErrTagNotFound):
		response.ErrorMsg(c, http.StatusNotFound, "tag not found")
	case errors.Is(err, constants.ErrCategoryNameEmpty):
		response.Error(c, http.StatusBadRequest, err)
	case errors.Is(err, constants.ErrCategoryFull):
		response.ErrorMsg(c, http.StatusConflict, fmt.Sprintf("this category already has %d words, the max per category", constants.MaxWordsPerCategory))
	default:
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
	}
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}
	view, err := h.svc.Create(c.Request.Context(), userID, req.Name)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, view)
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, gin.H{"items": items})
}

func (h *Handler) Delete(c *gin.Context) {
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
	if err := h.svc.Delete(c.Request.Context(), userID, id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) Tag(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	vocabID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid vocab id")
		return
	}
	var req struct {
		CategoryID uint64 `json:"categoryId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.Tag(c.Request.Context(), userID, vocabID, req.CategoryID); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func (h *Handler) Untag(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	vocabID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid vocab id")
		return
	}
	categoryID, err := strconv.ParseUint(c.Param("categoryId"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid category id")
		return
	}
	if err := h.svc.Untag(c.Request.Context(), userID, vocabID, categoryID); err != nil {
		writeServiceError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) ListForVocab(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	vocabID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid vocab id")
		return
	}
	items, err := h.svc.ListForVocab(c.Request.Context(), userID, vocabID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, gin.H{"items": items})
}
