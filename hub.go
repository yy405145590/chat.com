package main

import (
	"container/list"
	"flag"
	"net/url"
	"time"

	gateclient "github.com/chat.com/gate/client"
)

const (
	_ = iota
	ClientLoginTime
)

const (
	history_count = 50
)

type HubCommand struct {
	Args   interface{}
	CmdID  int
	Result chan interface{}
}

type Hub struct {
	clients   map[string]*Client
	broadcast chan []byte

	register   chan *Client
	unregister chan *Client
	C          chan *HubCommand
	msgHistory *list.List
	gateClient *gateclient.WsClient
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		C:          make(chan *HubCommand),
		msgHistory: list.New(),
	}
}

func (h *Hub) addHistory(message []byte) {
	h.msgHistory.PushFront(message)
	if h.msgHistory.Len() > history_count {
		h.msgHistory.Remove(h.msgHistory.Back())
	}
}

func (h *Hub) registerClient(client *Client) {
	_, ok := h.clients[client.username]
	if ok {
		client.stop <- 1
	} else {
		h.clients[client.username] = client
		for item := h.msgHistory.Back(); nil != item; item = item.Prev() {
			client.send <- item.Value.([]byte)
		}
	}
}

func (h *Hub) run() {
	var addr = flag.String("gateaddr", "localhost:8081", "gate service address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	h.gateClient = gateclient.NewWsClient()
	h.gateClient.Connect(u)
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			if _, ok := h.clients[client.username]; ok {
				delete(h.clients, client.username)
				close(client.send)
			}
		case message := <-h.broadcast:
			h.broadcastMsg(message)
			h.sendToGate(message)
		case cmd := <-h.C:
			h.dispatchCmd(cmd)
		case gateMsg := <-h.gateClient.R:
			h.broadcastMsg(gateMsg)
		}
	}
}

func (h *Hub) sendToGate(message []byte) {
	if h.gateClient.IsConnected {
		ticker := time.NewTicker(time.Second)
		select {
		case h.gateClient.S <- message:
		case <-ticker.C:
		}

	}
}

func (h *Hub) broadcastMsg(message []byte) {
	h.addHistory(message)
	for _, client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client.username)
		}
	}
}

func (h *Hub) getLoginTime(username string) int64 {
	client, ok := h.clients[username]
	if ok {
		return client.loginTime.Unix()
	}
	return 0
}

func (h *Hub) dispatchCmd(cmd *HubCommand) {
	switch cmd.CmdID {
	case ClientLoginTime:
		username := cmd.Args.(string)
		loginTs := h.getLoginTime(username)
		cmd.Result <- loginTs
	}
}
