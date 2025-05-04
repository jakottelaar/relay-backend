package websocket

import (
	"sync"

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
}

func (m *Manager) handleDisconnect(s *melody.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if userID, exists := m.sessionUsers[s]; exists {
		sessions := m.userSessions[userID]
		for i, session := range sessions {
			if session == s {
				// Remove this session
				m.userSessions[userID] = append(sessions[:i], sessions[i+1:]...)
				break
			}
		}

		// If no more sessions for this user, clean up
		if len(m.userSessions[userID]) == 0 {
			delete(m.userSessions, userID)
		}

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
