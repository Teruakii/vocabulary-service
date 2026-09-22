package stats

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/infra/http/response"
	"vocabulary-service/internal/middleware"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Me(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	s, err := h.service.Get(c.Request.Context(), userID)
	if err != nil {
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, s)
}
