package handlers

import (
    "net/http"

    "go-market/internal/ws"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

func WebSocketHandler(manager *ws.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDRaw, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
            return
        }
        userID := userIDRaw.(uint)

        conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
        if err != nil {
            return
        }
        client := &ws.Client{UserID: userID, Conn: conn}
        manager.AddClient(userID, client)
        defer func() {
            manager.RemoveClient(userID, client)
            conn.Close()
        }()

        // Keep the connection alive
        for {
            _, _, err := conn.ReadMessage()
            if err != nil {
                break
            }
        }
    }
}
