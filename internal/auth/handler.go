package auth

import (
	"church-match-api/pkg/middleware"
	"church-match-api/pkg/response"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Register(r.Context(), req); err != nil {
		if err.Error() == "email already taken" {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	response.JSON(w, http.StatusCreated, nil, "Registration successful. Your account is under review.")
}

func (h *Handler) RegisterAdmin(w http.ResponseWriter, r *http.Request) {
	var req AdminRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.RegisterAdmin(r.Context(), req, req.Secret); err != nil {
		switch err.Error() {
		case "invalid admin secret":
			response.Error(w, http.StatusForbidden, "Invalid admin secret key")
		case "email already taken":
			response.Error(w, http.StatusConflict, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "Failed to create admin account")
		}
		return
	}

	response.JSON(w, http.StatusCreated, nil, "Admin account created successfully. You may now log in.")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.service.Login(r.Context(), req)
	if err != nil {
		switch err.Error() {
		case "invalid credentials":
			response.Error(w, http.StatusUnauthorized, err.Error())
		default:
			response.Error(w, http.StatusForbidden, err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, res, "Login successful")
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	if err := h.service.Logout(r.Context(), userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Logout failed")
		return
	}
	response.JSON(w, http.StatusOK, nil, "Logged out successfully")
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	user, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch user data")
		return
	}
	if user == nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}
	response.JSON(w, http.StatusOK, user, "User data fetched successfully")
}
