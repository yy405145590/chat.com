package tools

import (
	"testing"
	"time"
)

func Test_popular(t *testing.T) {
	go P.Serve()
	testWord := map[string]int{
		"a": 10,
		"b": 20,
		"c": 30,
		"d": 40,
		"e": 50,
		"f": 60,
		"g": 70,
	}
	expected := map[string]bool{
		"c": true, "d": true, "e": true, "f": true, "g": true,
	}
	for k, v := range testWord {
		for i := 0; i < v; i++ {
			P.C <- &PopularCommand{
				Args: &MsgItem{
					Ts:      time.Now().Unix(),
					Message: k,
				},
				CmdID: CmdPutMsgItem,
			}
		}
	}
	resultC := make(chan interface{})
	P.C <- &PopularCommand{
		CmdID:  CmdGetPopularItems,
		Result: resultC,
	}
	ticker := time.NewTicker(5 * time.Second)
	select {
	case result := <-resultC:
		resMessage := result.([]string)
		for _, v := range resMessage {
			_, ok := expected[v]
			if !ok {
				t.Fail()
			}
		}
	case <-ticker.C:
		t.Fail()
	}
}
