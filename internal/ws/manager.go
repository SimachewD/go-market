package ws

import (
    "sync"

    "github.com/gorilla/websocket"
)

type Client struct {
    UserID uint
    Conn   *websocket.Conn
}

type Manager struct {
    clients map[uint]map[*Client]struct{} // userID -> set of clients
    mu      sync.RWMutex
}

func NewManager() *Manager {
    return &Manager{
        clients: make(map[uint]map[*Client]struct{}),
    }
}

func (m *Manager) AddClient(userID uint, client *Client) {
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.clients[userID] == nil {
        m.clients[userID] = make(map[*Client]struct{})
    }
    m.clients[userID][client] = struct{}{}
}

func (m *Manager) RemoveClient(userID uint, client *Client) {
    m.mu.Lock()
    defer m.mu.Unlock()
    if _, ok := m.clients[userID]; ok {
        delete(m.clients[userID], client)
        if len(m.clients[userID]) == 0 {
            delete(m.clients, userID)
        }
    }
}

func (m *Manager) SendToUser(userID uint, message interface{}) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    for client := range m.clients[userID] {
        client.Conn.WriteJSON(message)
    }
}
