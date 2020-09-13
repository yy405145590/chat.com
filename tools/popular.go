package tools

import (
	"time"
)

const (
	//popular word expire time
	expire_ts = 5
	//popular word expire time
	check_expire_ts      = time.Second
	populr_message_count = 5
)

//command popular
const (
	_ = iota
	CmdPutMsgItem
	CmdGetPopularItems
)

type MsgItem struct {
	Ts      int64
	Message string
}

type PopularCommand struct {
	Args   interface{}
	CmdID  int
	Result chan interface{}
}

type PopluarWord struct {
	C     chan *PopularCommand
	tmMap map[int64][]*MsgItem
}

func newPopluarWord() *PopluarWord {
	return &PopluarWord{C: make(chan *PopularCommand), tmMap: make(map[int64][]*MsgItem)}
}

func (p *PopluarWord) Serve() {
	ticker := time.NewTicker(check_expire_ts)
	for {
		select {
		case cmd := <-p.C:
			p.dispatch(cmd)
		case <-ticker.C:
			p.updateExpire()
		}
	}
}

func (p *PopluarWord) dispatch(cmd *PopularCommand) {
	switch cmd.CmdID {
	case CmdPutMsgItem:
		p.add(cmd.Args.(*MsgItem))
	case CmdGetPopularItems:
		poplrItems := p.getPopularItems()
		cmd.Result <- poplrItems
	}
}

func (p *PopluarWord) getPopularItems() []string {
	var messagesMap map[string]int = make(map[string]int)
	for _, vec := range p.tmMap {
		for _, val := range vec {
			messagesMap[val.Message]++
		}
	}
	sortedList := sortMapByValue(messagesMap, SortDESC)
	retList := make([]string, 0, populr_message_count)
	count := 0
	for _, v := range sortedList {
		retList = append(retList, v.Key)
		count++
		if count >= populr_message_count {
			break
		}
	}
	return retList
}

func (p *PopluarWord) add(msgItem *MsgItem) {
	v, ok := p.tmMap[msgItem.Ts]
	if !ok {
		v = make([]*MsgItem, 0, 10)

	}
	p.tmMap[msgItem.Ts] = append(v, msgItem)
}

func (p *PopluarWord) updateExpire() {
	now := time.Now().Unix()
	for k, _ := range p.tmMap {
		if k < now-int64(expire_ts) {
			delete(p.tmMap, k)
		}
	}
}

var P *PopluarWord = newPopluarWord()
