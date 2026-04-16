package admin

import (
	"church-match-api/pkg/response"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	users, err := h.service.GetUsers(r.Context(), status, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	response.JSON(w, http.StatusOK, users, "Users fetched successfully")
}

func (h *Handler) ApproveUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.ApproveUser(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to approve user")
		return
	}

	response.JSON(w, http.StatusOK, nil, "User approved successfully")
}

func (h *Handler) RejectUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.RejectUser(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to reject user")
		return
	}

	response.JSON(w, http.StatusOK, nil, "User rejected successfully")
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	response.JSON(w, http.StatusOK, nil, "User deleted successfully")
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch stats")
		return
	}

	response.JSON(w, http.StatusOK, stats, "Stats fetched successfully")
}
