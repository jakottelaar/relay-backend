package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"slices"

	"github.com/google/uuid"
	"github.com/olahol/melody"
)

type Manager struct {
	melody       *melody.Melody
	userSessions map[uuid.UUID][]*melody.Session
	sessionUsers map[*melody.Session]uuid.UUID
	mu           sync.RWMutex
}

func NewManager() *Manager {
	m := melody.New()

	manager := &Manager{
		melody:       m,
		userSessions: make(map[uuid.UUID][]*melody.Session),
		sessionUsers: make(map[*melody.Session]uuid.UUID),
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

	m.userSessions[userID] = append(m.userSessions[userID], s)
	m.sessionUsers[s] = userID

	resp := struct {
		Type string `json:"type"`
		Data string `json:"data"`
	}{
		Type: "USER_WENT_ONLINE",
		Data: userID.String(),
	}

	data, err := json.Marshal(resp)
	if err == nil {
		m.melody.Broadcast(data)
	}
}

func (m *Manager) handleDisconnect(s *melody.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if userID, exists := m.sessionUsers[s]; exists {
		sessions := m.userSessions[userID]
		for i, session := range sessions {
			if session == s {
				// Remove this session
				m.userSessions[userID] = slices.Delete(sessions, i, i+1)
				break
			}
		}

		// If no more sessions for this user, clean up
		go func(userID uuid.UUID) {
			time.Sleep(5 * time.Second) // Wait a bit before checking if the user is still online

			m.mu.Lock()
			defer m.mu.Unlock()

			sessions := m.userSessions[userID]
			if len(sessions) == 0 {
				// Now they're *really* offline
				payload := struct {
					Type string `json:"type"`
					Data string `json:"data"`
				}{
					Type: "USER_WENT_OFFLINE",
					Data: userID.String(),
				}

				data, _ := json.Marshal(payload)
				m.melody.Broadcast(data)
			}
		}(userID)

		delete(m.sessionUsers, s)
	}
}

// SendToUser sends a message to all sessions of a specific user
func (m *Manager) SendToUser(userID uuid.UUID, message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions, exists := m.userSessions[userID]
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
	}
}

func (m *Manager) handleGetOnlineUsers(s *melody.Session) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var onlineUserIDs []string
	for userID := range m.userSessions {
		onlineUserIDs = append(onlineUserIDs, userID.String())
	}

	resp := struct {
		Type string   `json:"type"`
		Data []string `json:"data"`
	}{
		Type: "ONLINE_USERS",
		Data: onlineUserIDs,
	}

	data, err := json.Marshal(resp)
	if err == nil {
		s.Write(data)
	}
}
