package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
)

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func main() {
	fmt.Println("Starting backend...")
	go manager.start()
	http.HandleFunc("/ws", wsPage)
	http.ListenAndServe(":12345", nil)
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
			msg, err := json.Marshal(&Message{
				Content: "A new client has connected.",
			})
			if err != nil {
				log.Println("JSON marshal fails. " + err.Error())
			}
			manager.send(msg, conn)
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				msg, err := json.Marshal(&Message{
					Content: "A client quits.",
				})
				if err != nil {
					log.Println("JSON marshal fails. " + err.Error())
				}
				manager.send(msg, conn)
			}
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				conn.send <- message
			}
		}
	}
}

func (manager *ClientManager) send(m []byte, ignore *Client) {
	for conn := range manager.clients {
		if conn != ignore {
			conn.send <- m
		}
	}
}

func (client *Client) read() {
	defer func() {
		manager.unregister <- client
		client.socket.Close()
	}()

	for {
		_, message, err := client.socket.ReadMessage()
		if err != nil {
			manager.unregister <- client
			client.socket.Close()
			break
		}

		msg, _ := json.Marshal(&Message{
			Sender:  client.id,
			Content: string(message),
		})
		manager.broadcast <- msg
	}
}

func (client *Client) write() {
	defer func() {
		client.socket.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				client.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			client.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func wsPage(w http.ResponseWriter, r *http.Request) {
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(rr *http.Request) bool { return true },
	}).Upgrade(w, r, nil)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	client := &Client{
		id:     uuid.NewV4().String(),
		socket: conn,
		send:   make(chan []byte),
	}

	manager.register <- client

	// fmt.Println("new client: " + client.id)

	go client.read()
	go client.write()
}
