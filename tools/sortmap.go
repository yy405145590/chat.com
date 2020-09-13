package tools

import (
	"fmt"
	"sort"
)

type ItemPair struct {
	Key   string
	Value int
}
type PairList []ItemPair

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value > p[j].Value }

const (
	_ = iota
	SortDESC
)

func sortMapByValue(m map[string]int, tye int) PairList {
	if tye != SortDESC {
		panic(fmt.Sprintf("tye:%v not implment", tye))
	}
	p := make(PairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = ItemPair{k, v}
		i++
	}
	sort.Sort(p)
	return p
}
