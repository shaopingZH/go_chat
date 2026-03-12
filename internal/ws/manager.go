package ws

import (
	"sync"
	"time"
)

type Manager struct {
	mu           sync.RWMutex
	clients      map[uint64]*Client
	lastSeen     map[uint64]time.Time
	RedisAdapter *RedisAdapter
}

func NewManager() *Manager {
	return &Manager{clients: make(map[uint64]*Client), lastSeen: make(map[uint64]time.Time)}
}

func (m *Manager) SetRedisAdapter(adapter *RedisAdapter) {
	m.RedisAdapter = adapter
}

func (m *Manager) Register(client *Client) {
	m.mu.Lock()
	old := m.clients[client.UserID]
	m.clients[client.UserID] = client
	m.mu.Unlock()

	if old != nil && old != client {
		old.Close()
	}
}

func (m *Manager) Unregister(client *Client) {
	m.mu.Lock()
	current := m.clients[client.UserID]
	if current == client {
		delete(m.clients, client.UserID)
		m.lastSeen[client.UserID] = time.Now().UTC()
	}
	m.mu.Unlock()

	client.Close()
}

func (m *Manager) SendToUser(userID uint64, payload []byte) bool {
	m.mu.RLock()
	client := m.clients[userID]
	m.mu.RUnlock()

	if client != nil {
		return client.Enqueue(payload)
	}

	return false
}

func (m *Manager) BroadcastAllLocal(payload []byte) {
	m.mu.RLock()
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.mu.RUnlock()

	for _, client := range clients {
		if client != nil {
			client.Enqueue(payload)
		}
	}
}

func (m *Manager) IsOnline(userID uint64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.clients[userID]
	return ok
}

func (m *Manager) LastSeen(userID uint64) (time.Time, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.lastSeen[userID]
	return value, ok
}

// Broadcast sends the message to the user, either locally or via Redis
func (m *Manager) Broadcast(userID uint64, payload []byte) error {
	// 1. Try local send first (optional optimization)
	if m.SendToUser(userID, payload) {
		return nil
	}

	// 2. If not local (or even if local, to support multi-device), publish to Redis
	if m.RedisAdapter != nil {
		return m.RedisAdapter.Publish(userID, payload)
	}

	return nil
}
