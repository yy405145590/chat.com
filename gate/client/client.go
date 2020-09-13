package client

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	c           *websocket.Conn
	S           chan []byte
	R           chan []byte
	done        chan struct{}
	IsConnected bool
}

func NewWsClient() *WsClient {
	return &WsClient{
		S: make(chan []byte, 256),
		R: make(chan []byte, 256),
	}
}

func (wsclient *WsClient) Connect(u url.URL) bool {
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("dial:%v", err)
		return false
	}
	wsclient.c = c
	wsclient.done = make(chan struct{})

	go wsclient.loopRead()
	go wsclient.loopWrite()
	wsclient.IsConnected = true
	log.Printf("connect to %s ok", u.String())
	return true
}

func (wsclient *WsClient) loopRead() {
	defer func() {
		wsclient.IsConnected = false
		close(wsclient.done)
	}()
	for {
		_, message, err := wsclient.c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		wsclient.R <- message
	}
}

func (wsClient *WsClient) loopWrite() {
	select {
	case message := <-wsClient.S:
		err := wsClient.c.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("write:", err)
			return
		}
	case <-wsClient.done:
		return
	}
}
