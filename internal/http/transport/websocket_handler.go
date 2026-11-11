package transport

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/Rizwank123/chatApp/internal/websockets"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWebSocket(hub *websockets.Hub, c echo.Context, personnelID string) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Could not open websocket connection")
	}

	client := &websockets.Client{
		ID:         personnelID,
		ReceiverID: c.QueryParam("receiver_id"),
		Conn:       conn,
		Send:       make(chan []byte, 256),
		Hub:        hub,
	}

	hub.Register <- client

	go client.WritePump()
	go client.ReadPump()

	return nil
}
