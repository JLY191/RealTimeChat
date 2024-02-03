package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
)

// a global manager
var manager = ClientManager{

	// init channels and map
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

// main goroutine
func main() {

	// prompt
	fmt.Println("Starting backend...")

	// ClientManager starts
	go manager.start()

	// http listen
	http.HandleFunc("/ws", wsPage)
	http.ListenAndServe(":12345", nil)
}

func (manager *ClientManager) start() {

	// listen three channels
	for {
		select {

		// new client
		case conn := <-manager.register:

			// add map
			manager.clients[conn] = true
			msg, err := json.Marshal(&Message{
				Content: "A new client has connected.",
			})
			if err != nil {
				log.Println("JSON marshal fails. " + err.Error())
			}

			// notify other clients
			manager.send(msg, conn)

		// disconnect
		case conn := <-manager.unregister:

			// delete map
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				msg, err := json.Marshal(&Message{
					Content: "A client quits.",
				})
				if err != nil {
					log.Println("JSON marshal fails. " + err.Error())
				}

				// notify other clients
				manager.send(msg, conn)
			}

		// a client wants to send message
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				conn.send <- message
			}
		}
	}
}

// send message to other clients
// used for server message
// like 'register' or 'disconnect'
func (manager *ClientManager) send(m []byte, ignore *Client) {
	for conn := range manager.clients {
		if conn != ignore {
			conn.send <- m
		}
	}
}

// read from client
func (client *Client) read() {

	// defer
	defer func() {
		manager.unregister <- client
		client.socket.Close()
	}()

	for {

		// read
		_, message, err := client.socket.ReadMessage()

		// read fail, stands for disconnected client
		if err != nil {
			manager.unregister <- client
			client.socket.Close()
			break
		}

		// send message
		msg, _ := json.Marshal(&Message{
			Sender:  client.id,
			Content: string(message),
		})
		manager.broadcast <- msg
	}
}

// write to user
func (client *Client) write() {

	// defer
	defer func() {
		client.socket.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:

			// send chan is closed, which means that ClientManager has closed websocket
			// so, notify user end to close the other side of websocket
			if !ok {
				client.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			client.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// websocket upgrader
func wsPage(w http.ResponseWriter, r *http.Request) {
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(rr *http.Request) bool { return true },
	}).Upgrade(w, r, nil)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	// client
	client := &Client{
		id:     uuid.NewV4().String(),
		socket: conn,
		send:   make(chan []byte),
	}

	// register
	manager.register <- client

	// start rw goroutines
	go client.read()
	go client.write()
}
