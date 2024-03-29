package main

// struct used in this mini program

import "github.com/gorilla/websocket"

// Client stands for a connection between a user and the server
type Client struct {

	// uuid, generated by the go.uuid lib
	id string

	// websocket
	socket *websocket.Conn

	// can be seen as a buffer containing message to be sent to user
	// when sending message to user, read from this chan and send through websocket api
	send chan []byte
}

// ClientManager is the master node
type ClientManager struct {

	// containing every client alive
	clients map[*Client]bool

	// message to be sent to every client alive
	broadcast chan []byte

	// listen on register
	register chan *Client

	// listen on unregister
	unregister chan *Client
}

// Message is a JSON object
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}
