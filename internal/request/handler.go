package request

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

func (h *Handler) SendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	senderID, _ := ctx.Value(middleware.UserIDKey).(string)

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.service.SendRequest(ctx, senderID, req.ReceiverID); err != nil {
		if err.Error() == "request already exists" {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, nil, "Request sent successfully")
}

func (h *Handler) GetReceived(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)

	requests, err := h.service.GetReceived(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch received requests")
		return
	}

	response.JSON(w, http.StatusOK, requests, "Received requests fetched successfully")
}

func (h *Handler) GetSent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)

	requests, err := h.service.GetSent(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch sent requests")
		return
	}

	response.JSON(w, http.StatusOK, requests, "Sent requests fetched successfully")
}

func (h *Handler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	id := vars["id"]

	chatID, err := h.service.AcceptRequest(ctx, userID, id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"chat_id": chatID}, "Request accepted")
}

func (h *Handler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.RejectRequest(ctx, userID, id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, nil, "Request rejected")
}

func (h *Handler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.CancelRequest(ctx, userID, id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, nil, "Request cancelled")
}
