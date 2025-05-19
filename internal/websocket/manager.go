package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"slices"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal/messages"
	"github.com/olahol/melody"
)

type EventPayload struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Manager struct {
	melody              *melody.Melody
	sessionsByUserID    map[uuid.UUID][]*melody.Session
	userIDBySession     map[*melody.Session]uuid.UUID
	sessionsByChannelID map[string]map[*melody.Session]struct{}
	mu                  sync.RWMutex
	messageService      messages.MessagesService
}

func NewManager(messagesService messages.MessagesService) *Manager {
	m := melody.New()

	manager := &Manager{
		melody:              m,
		sessionsByUserID:    make(map[uuid.UUID][]*melody.Session),
		userIDBySession:     make(map[*melody.Session]uuid.UUID),
		sessionsByChannelID: make(map[string]map[*melody.Session]struct{}),
		messageService:      messagesService,
	}

	// Set up handlers
	m.HandleConnect(manager.handleConnect)
	m.HandleDisconnect(manager.handleDisconnect)

	m.HandleMessage(manager.handleMessage)

	return manager
}

func (m *Manager) handleConnect(s *melody.Session) {
	// Auth is handled at the HTTP handler level before upgrading to WebSocket
	userIDValue, exists := s.Get("user_id")
	if !exists {
		return
	}
	userID, err := uuid.Parse(userIDValue.(string))
	if err != nil {
		s.Close()
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessionsByUserID[userID] = append(m.sessionsByUserID[userID], s)
	m.userIDBySession[s] = userID

	payload := &EventPayload{
		Type: "USER_WENT_ONLINE",
		Data: userID.String(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	m.melody.Broadcast(data)

}

func (m *Manager) handleDisconnect(s *melody.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if userID, exists := m.userIDBySession[s]; exists {
		sessions := m.sessionsByUserID[userID]
		for i, session := range sessions {
			if session == s {
				// Remove this session
				m.sessionsByUserID[userID] = slices.Delete(sessions, i, i+1)
				break
			}
		}

		// If no more sessions for this user, clean up
		go func(userID uuid.UUID) {
			time.Sleep(5 * time.Second) // Wait a bit before checking if the user is still online

			m.mu.Lock()
			defer m.mu.Unlock()

			sessions := m.sessionsByUserID[userID]
			if len(sessions) == 0 {
				// Now they're *really* offline
				payload := &EventPayload{
					Type: "USER_WENT_OFFLINE",
					Data: userID.String(),
				}

				data, _ := json.Marshal(payload)
				m.melody.Broadcast(data)
			}
		}(userID)

		delete(m.userIDBySession, s)
	}
}

// SendToUser sends a message to all sessions of a specific user
func (m *Manager) SendToUser(userID uuid.UUID, message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions, exists := m.sessionsByUserID[userID]
	if exists {
		for _, session := range sessions {
			session.Write(message)
		}
	}
}

// GetMelody returns the underlying melody instance for HTTP handler registration
func (m *Manager) GetMelody() *melody.Melody {
	return m.melody
}

func (m *Manager) sendError(s *melody.Session, errorMessage string) {
	resp := struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}{
		Type:    "ERROR",
		Message: errorMessage,
	}

	data, err := json.Marshal(resp)
	if err == nil {
		s.Write(data)
	}
}

func (m *Manager) handleMessage(s *melody.Session, msg []byte) {
	var incoming struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(msg, &incoming); err != nil {
		return
	}

	switch incoming.Type {
	case "GET_ONLINE_USERS":
		m.handleGetOnlineUsers(s)

	case "SEND_MESSAGE":
		m.handleSendMessage(s, msg)
	case "JOIN_CHANNEL":

		var payload struct {
			ChannelID string `json:"channel_id"`
		}

		if err := json.Unmarshal(msg, &payload); err != nil {
			return
		}

		m.JoinChannel(payload.ChannelID, s)

	}
}

func (m *Manager) handleGetOnlineUsers(s *melody.Session) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var onlineUserIDs []string
	for userID := range m.sessionsByUserID {
		onlineUserIDs = append(onlineUserIDs, userID.String())
	}

	payload := &EventPayload{
		Type: "ONLINE_USERS",
		Data: onlineUserIDs,
	}

	data, err := json.Marshal(payload)
	if err == nil {
		s.Write(data)
	}
}

func (m *Manager) JoinChannel(channelID string, s *melody.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sessionsByChannelID == nil {
		m.sessionsByChannelID = make(map[string]map[*melody.Session]struct{})
	}

	if m.sessionsByChannelID[channelID] == nil {
		m.sessionsByChannelID[channelID] = make(map[*melody.Session]struct{})
	}

	m.sessionsByChannelID[channelID][s] = struct{}{}
	s.Set("channel_id", channelID)

	log.Printf("joined channel %s", channelID)
}

func (m *Manager) LeaveChannel(channelID string, s *melody.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sessions, ok := m.sessionsByChannelID[channelID]; ok {
		delete(sessions, s)
		if len(sessions) == 0 {
			delete(m.sessionsByChannelID, channelID)
		}
	}
}

func (m *Manager) BroadcastToChannel(channelID string, data []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if sessions, ok := m.sessionsByChannelID[channelID]; ok {
		for s := range sessions {
			s.Write(data)
		}
	}
}

func (m *Manager) sessionInChannel(s *melody.Session, channelID string) bool {
	id, ok := s.Get("channel_id")
	return ok && id == channelID
}

func (m *Manager) handleSendMessage(s *melody.Session, msg []byte) {
	var payload struct {
		ChannelID string `json:"channel_id"`
		Content   string `json:"content"`
	}

	if err := json.Unmarshal(msg, &payload); err != nil {
		m.sendError(s, "Invalid message format")
		return
	}

	currentUserID, ok := s.Get("user_id")
	if !ok {
		m.sendError(s, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		m.sendError(s, "Invalid user ID")
		return
	}

	message, err := m.messageService.CreateMessage(context.Background(), userID, uuid.MustParse(payload.ChannelID), payload.Content)
	if err != nil {
		m.sendError(s, "Failed to save message")
		return
	}

	resp := struct {
		Type    string                          `json:"type"`
		Message *messages.CreateMessageResponse `json:"message"`
		Sender  uuid.UUID                       `json:"sender"`
	}{
		Type: "MESSAGE_SENT",
		Message: &messages.CreateMessageResponse{
			ID:        message.ID,
			SenderID:  message.SenderID,
			ChannelID: message.ChannelID,
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
		},
		Sender: message.SenderID,
	}

	log.Printf("Broadcasting message to channel %s: %+v", payload.ChannelID, resp)

	responseJSON, _ := json.Marshal(resp)
	m.melody.BroadcastFilter(responseJSON, func(s *melody.Session) bool {
		return m.sessionInChannel(s, payload.ChannelID)
	})

}
