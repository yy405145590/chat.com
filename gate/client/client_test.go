package client

import (
	"flag"
	"net/url"
	"testing"
	"time"

	"github.com/chat.com/gate/server"
)

func Test_Client(t *testing.T) {
	go server.Run()
	var addr = flag.String("gateaddr", "localhost:8081", "gate service address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	clients := make([]*WsClient, 2)
	for i := 0; i < len(clients); i++ {
		clients[i] = NewWsClient()
		clients[i].Connect(u)
		if !clients[i].IsConnected {
			t.Fail()
		}
	}
	for i := 0; i < len(clients); i++ {
		clients[i].S <- []byte("1")
	}
	for i := 0; i < len(clients); i++ {
		ticker := time.NewTicker(time.Second * 2)
		select {
		case msg := <-clients[i].R:
			if string(msg) != "1" {
				t.Fail()
			}
		case <-ticker.C:
			t.Fail()
		}
	}

}
