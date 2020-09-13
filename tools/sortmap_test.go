package tools

import (
	"fmt"
	"testing"
)

func TestSortMap(t *testing.T) {
	m := make(map[string]int)
	m["a"] = 10
	m["b"] = 20
	m["c"] = 15
	r := sortMapByValue(m, SortDESC)
	fmt.Print(r)
}
