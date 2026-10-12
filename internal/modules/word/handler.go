package word

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

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, constants.ErrWordNotFound):
		response.ErrorMsg(c, http.StatusNotFound, "word not found")
	case errors.Is(err, constants.ErrWordDuplicate):
		response.ErrorMsg(c, http.StatusConflict, "word already exists")
	case errors.Is(err, constants.ErrWordInUse):
		response.ErrorMsg(c, http.StatusConflict, "word is in use and cannot be deleted")
	case errors.Is(err, constants.ErrWordForbidden):
		response.ErrorMsg(c, http.StatusForbidden, "you don't have permission to modify this word")
	case errors.Is(err, constants.ErrWordEmpty), errors.Is(err, constants.ErrMeaningEmpty):
		response.Error(c, http.StatusBadRequest, err)
	default:
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
	}
}

func callerInfo(c *gin.Context) (userID uint64, isAdmin bool, ok bool) {
	userID, ok = middleware.CurrentUserID(c)
	if !ok {
		return 0, false, false
	}
	role, _ := c.Get(middleware.ContextKeyRole)
	isAdmin = role == string(constants.RoleAdmin)
	return userID, isAdmin, true
}

func (h *Handler) Create(c *gin.Context) {
	userID, isAdmin, ok := callerInfo(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	var owner *uint64
	if !(req.Global && isAdmin) {
		owner = &userID
	}

	w, err := h.service.AddWord(c.Request.Context(), owner, req.Word, req.Meaning)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, w)
}

func (h *Handler) GetByID(c *gin.Context) {
	userID, _, ok := callerInfo(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid id")
		return
	}

	w, err := h.service.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, w)
}

func (h *Handler) GetByWord(c *gin.Context) {
	userID, _, ok := callerInfo(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	wordStr := c.Query("word")
	if wordStr == "" {
		response.ErrorMsg(c, http.StatusBadRequest, "word query param is required")
		return
	}

	w, err := h.service.GetByWord(c.Request.Context(), userID, wordStr)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, w)
}

func (h *Handler) GetAll(c *gin.Context) {
	userID, _, ok := callerInfo(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	words, hasMore, err := h.service.GetAllWord(c.Request.Context(), userID, limit, offset)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, gin.H{"items": words, "hasMore": hasMore})
}

func (h *Handler) Update(c *gin.Context) {
	userID, isAdmin, ok := callerInfo(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdateWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	w := &Word{ID: id}
	if req.Word != nil {
		w.Word = *req.Word
	}
	if req.Meaning != nil {
		w.Meaning = *req.Meaning
	}

	if err := h.service.UpdateWord(c.Request.Context(), userID, isAdmin, w); err != nil {
		writeServiceError(c, err)
		return
	}
	updated, err := h.service.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, http.StatusOK, updated)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, isAdmin, ok := callerInfo(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.Delete(c.Request.Context(), userID, isAdmin, id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.NoContent(c)
}
