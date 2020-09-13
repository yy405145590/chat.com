package server

import (
	"flag"
	"net/url"
	"testing"
	"time"

	"github.com/chat.com/gate/client"
)

func Test_Client(t *testing.T) {
	go Run()
	var addr = flag.String("gateaddr", "localhost:8081", "gate service address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	clients := make([]*client.WsClient, 2)
	for i := 0; i < len(clients); i++ {
		clients[i] = client.NewWsClient()
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
