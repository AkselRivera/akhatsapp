package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MAX_MESSAGE_SIZE = 512
	PING_PERIOD      = 10 * time.Minute
	PONG_PERIOD      = PING_PERIOD + (10 * time.Second)
	WRITE_WAIT       = 10 * time.Second // Maximum time to wait before clients should send message
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	Nickname string `json:"nickname,omitempty"`
	Content  string `json:"content,omitempty"`
}

type Client struct {
	nickname     string
	hub          *Hub
	conn         *websocket.Conn
	queueMessage chan Message // Buffered channel
}

func (c *Client) readWS() {

	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(MAX_MESSAGE_SIZE)
	c.conn.SetReadDeadline(time.Now().Add(PONG_PERIOD))

	c.conn.SetPongHandler(func(ping string) error {
		fmt.Println(ping, "Pong: ", c.nickname)

		c.conn.SetReadDeadline(time.Now().Add(PING_PERIOD))
		return nil
	})

	for {
		var message Message
		if err := c.conn.ReadJSON(&message); err != nil {

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Error cannot read WS message: ", err)
			}

			return
		}

		c.hub.broadcast <- message

	}
}

func (c *Client) writeWS() {
	ticker := time.NewTicker(PING_PERIOD)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, isOpen := <-c.queueMessage:

			if !isOpen {
				// return if channel is closed by hub
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

			if err := c.conn.WriteJSON(message); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("Error cannot write WS message: ", err)
					return
				}
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

			if err := c.conn.WriteMessage(websocket.PingMessage, []byte("Ping")); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("Error cannot write Ping message: ", err)
					return
				}
			}

		}
	}
}

func wsHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {

	nickname := r.URL.Query()["nickname"]

	if len(nickname) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please provide a nickname"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Error upgrading WS connection: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	client := &Client{
		nickname:     nickname[0],
		hub:          hub,
		conn:         conn,
		queueMessage: make(chan Message, 5),
	}

	client.hub.register <- client

	go client.writeWS()
	go client.readWS()
}
