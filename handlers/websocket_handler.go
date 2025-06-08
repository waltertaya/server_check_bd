package handlers

import (
	"encoding/json"
	"log"

	"github.com/waltertaya/server_check_bd/services"
	"golang.org/x/net/websocket"
)

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(checker *services.HealthChecker) websocket.Handler {
	return func(ws *websocket.Conn) {
		client := &services.Client{
			Send: make(chan interface{}, 256),
		}

		// Register client
		checker.RegisterClient(client)
		defer checker.UnregisterClient(client)

		// Start goroutine to send messages to client
		go func() {
			for message := range client.Send {
				data, err := json.Marshal(message)
				if err != nil {
					log.Printf("Error marshaling message: %v", err)
					continue
				}

				if err := websocket.Message.Send(ws, string(data)); err != nil {
					log.Printf("Error sending message: %v", err)
					return
				}
			}
		}()

		// Handle incoming messages
		for {
			var message string
			err := websocket.Message.Receive(ws, &message)
			if err != nil {
				if err.Error() != "EOF" {
					log.Printf("Error receiving message: %v", err)
				}
				break
			}

			// Parse message
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(message), &data); err != nil {
				log.Printf("Error parsing message: %v", err)
				continue
			}

			// Handle message based on type
			if msgType, ok := data["type"].(string); ok {
				switch msgType {
				case "check:server":
					if id, ok := data["id"].(string); ok {
						checker.RequestCheckServer(id)
					}
				default:
					log.Printf("Unknown message type: %s", msgType)
				}
			}
		}
	}
}
