package websockets

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Message struct {
	Type      string      `json:"type"`
	Content   interface{} `json:"content"`
	SenderID  string      `json:"sender_id"`
	Timestamp time.Time   `json:"timestamp"`
}

type Client struct {
	ID         string
	ReceiverID string
	Conn       *websocket.Conn
	Send       chan []byte
	Hub        *Hub
}

func (c *Client) ReadPump() {
	defer func() {
		log.Printf("Client %s disconnecting from ReadPump", c.ID)
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Read error for client %s: %v", c.ID, err)
			break
		}

		log.Printf("Received message from client %s: %s", c.ID, string(message))

		// Parse the message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error parsing message from client %s: %v", c.ID, err)
			continue
		}

		// Add sender ID and timestamp
		msg.SenderID = c.ID
		msg.Timestamp = time.Now()

		// Broadcast to specific receiver
		log.Printf("Sending message from %s to %s", c.ID, c.ReceiverID)
		c.Hub.Broadcast <- MessageEnvelope{
			SenderID:   c.ID,
			ReceiverID: c.ReceiverID,
			Message:    message,
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Printf("Client %s disconnecting from WritePump", c.ID)
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Printf("Channel closed for client %s", c.ID)
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Error getting writer for client %s: %v", c.ID, err)
				return
			}
			log.Printf("Writing message to client %s: %s", c.ID, string(message))
			w.Write(message)
			w.Close()

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping to client %s: %v", c.ID, err)
				return
			}
		}
	}
}
