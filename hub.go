package main

type Hub struct {
	clients    map[string]*Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {

	for {
		select {
		case client := <-h.register:
			h.clients[client.nickname] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.nickname]; ok {
				delete(h.clients, client.nickname)
				close(client.queueMessage)
			}

		case message := <-h.broadcast:
			for nickname, client := range h.clients {
				if nickname != message.Nickname {
					select {
					case client.queueMessage <- message:

					default:
						close(client.queueMessage)
						delete(h.clients, nickname)
					}
				}
			}
		}
	}
}
