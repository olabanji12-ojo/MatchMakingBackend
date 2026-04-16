package profile

import (
	"church-match-api/pkg/middleware"
	"church-match-api/pkg/response"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	service  Service
	validate *validator.Validate
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *Handler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)

	profile, err := h.service.GetProfile(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch profile")
		return
	}

	// Return an empty profile shell for new users so the frontend can show a blank form
	if profile == nil {
		profile = &Profile{}
	}

	response.JSON(w, http.StatusOK, profile, "Profile fetched successfully")
}

func (h *Handler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)

	var prof Profile
	if err := json.NewDecoder(r.Body).Decode(&prof); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(prof); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.service.UpdateProfile(ctx, userID, prof)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	response.JSON(w, http.StatusOK, updated, "Profile updated successfully")
}

func (h *Handler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	profile, err := h.service.GetPublicProfile(ctx, id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch public profile")
		return
	}

	if profile == nil {
		response.Error(w, http.StatusNotFound, "Profile not found")
		return
	}

	response.JSON(w, http.StatusOK, profile, "Public profile fetched successfully")
}
