package match

import (
	"church-match-api/pkg/middleware"
	"church-match-api/pkg/response"
	"net/http"
	"strconv"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	matches, err := h.service.GetMatches(ctx, userID, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch matches")
		return
	}

	response.JSON(w, http.StatusOK, matches, "Matches fetched successfully")
}
