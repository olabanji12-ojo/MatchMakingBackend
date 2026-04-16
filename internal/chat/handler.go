package chat

import (
	"church-match-api/pkg/middleware"
	"church-match-api/pkg/response"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, restrict this to your frontend domain
	},
}

type Handler struct {
	service  Service
	hub      *Hub
	validate *validator.Validate
}

func NewHandler(service Service, hub *Hub) *Handler {
	return &Handler{
		service:  service,
		hub:      hub,
		validate: validator.New(),
	}
}

func (h *Handler) GetChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)

	summaries, err := h.service.GetChatList(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch chats")
		return
	}

	response.JSON(w, http.StatusOK, summaries, "Chats fetched successfully")
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	chatID := vars["id"]

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	before := r.URL.Query().Get("before")

	messages, err := h.service.GetMessages(ctx, userID, chatID, limit, before)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, messages, "Messages fetched successfully")
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	chatID := vars["id"]

	var payload struct {
		Content string `json:"content" validate:"required,max=2000"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	msg, err := h.service.SendMessage(ctx, userID, chatID, payload.Content)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	// Broadast via WS to the other user
	otherUserID, _ := h.service.VerifyParticipation(ctx, userID, chatID)
	h.hub.BroadcastToUser(otherUserID, WSMessage{
		Type:    "new_message",
		ChatID:  chatID,
		Message: *msg,
	})

	response.JSON(w, http.StatusCreated, msg, "Message sent")
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(middleware.UserIDKey).(string)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	client := &Client{
		userID: userID,
		hub:    h.hub,
		conn:   conn,
		send:   make(chan []byte, 256),
	}

	client.hub.register <- client

	// Start pumps
	go client.WritePump()
	go client.ReadPump(h.service)
}
