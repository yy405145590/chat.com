package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chat.com/tools"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	stop      chan int
	loginTime time.Time
	username  string
	hasReg    bool
}

func (c *Client) handleReg(message []byte) bool {
	arr := strings.Split(string(message), "|")
	if len(arr) != 2 {
		return false
	}
	if arr[0] != "reg" {
		return false
	}
	c.hasReg = true
	c.username = arr[1]
	c.hub.register <- c
	return true
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		if !c.hasReg {
			if !c.handleReg(message) {
				log.Printf("login fail")
				break
			}
			continue
		}
		if !c.checkGmMessge(string(message)) {
			message = tools.FilterMessage(message)
			message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
			tools.P.C <- &tools.PopularCommand{
				Args: &tools.MsgItem{
					Ts:      time.Now().Unix(),
					Message: string(message),
				},
				CmdID: tools.CmdPutMsgItem,
			}
			c.hub.broadcast <- append([]byte(c.username+":"), message...)
		}
	}
}

func (c *Client) checkGmMessge(message string) bool {
	if len(message) <= 0 {
		return false
	}
	if message[0] == '/' {
		arr := strings.Split(message, " ")
		if len(arr) <= 0 {
			return false
		}
		switch arr[0] {
		case "/popular":
			{
				resultC := make(chan interface{})
				tools.P.C <- &tools.PopularCommand{
					CmdID:  tools.CmdGetPopularItems,
					Result: resultC,
				}
				ticker := time.NewTicker(5 * time.Second)
				select {
				case result := <-resultC:
					resMessage := result.([]string)
					for _, v := range resMessage {
						c.send <- []byte(v)
					}
				case <-ticker.C:
				}
			}
		case "/stats":
			loginTs := int64(0)
			if len(arr) > 1 {
				resultC := make(chan interface{}, 1)
				c.hub.C <- &HubCommand{
					CmdID:  ClientLoginTime,
					Args:   arr[1],
					Result: resultC,
				}
				tmpTs := <-resultC
				loginTs = tmpTs.(int64)
			} else {
				loginTs = c.loginTime.Unix()
			}
			msg := ""
			if loginTs == 0 {
				msg = "user not login"
			} else {
				diff := time.Now().Unix() - loginTs
				var days = diff / (24 * 60 * 60)
				diff = diff % (24 * 60 * 60)
				var hour = diff / (60 * 60)
				diff = diff % (60 * 60)
				var m = diff / 60
				diff = diff % 60
				msg = fmt.Sprintf("%03dd %02dh %02dm %02ds", days, hour, m, diff)
			}
			c.send <- []byte(msg)
		default:
			c.send <- []byte("Unrecognized Command")
		}
		return true
	}
	return false
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case stopCode := <-c.stop:
			log.Printf("stopcode:%v", stopCode)
			return
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), stop: make(chan int, 1), loginTime: time.Now()}
	go client.writePump()
	go client.readPump()
}
