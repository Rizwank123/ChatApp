package transport

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/Rizwank123/chatApp/internal/websockets"
)

func ServeWebSocket(hub *websockets.Hub, c echo.Context, personnelID string, allowedOrigins []string) error {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				return true // Native clients do not send an Origin header.
			}
			parsedOrigin, err := url.Parse(origin)
			if err != nil || parsedOrigin.Scheme == "" || parsedOrigin.Host == "" {
				return false
			}
			for _, allowed := range allowedOrigins {
				if strings.EqualFold(origin, strings.TrimSpace(allowed)) {
					return true
				}
			}
			return false
		},
	}
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Could not open websocket connection")
	}

	client := &websockets.Client{
		ID:   personnelID,
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  hub,
	}

	if err := hub.RegisterClient(client); err != nil {
		_ = conn.Close()
		return echo.NewHTTPError(http.StatusServiceUnavailable, "WebSocket server is shutting down")
	}

	go client.WritePump()
	go client.ReadPump()

	return nil
}
