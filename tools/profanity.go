package tools

import (
	"bufio"
	"os"
	"strings"
)

var (
	profanityTree = NewRoot()
)

func loadProfanity(filename string) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		profanityTree.Insert(line)
		if err != nil {
			break
		}
	}
	return true
}

func init() {
	loadProfanity("word.txt")
}

const max_length = 100

func FilterMessage(message []byte) []byte {
	if len(message) > max_length {
		return []byte("message exceed 1000 length")
	}

	lastMatchWord := ""
	tmp := ""
	for _, v := range message {
		tmp = tmp + string(v)
		if !profanityTree.Search(tmp) {
			continue
		}
		lastMatchWord = tmp
	}
	for i, _ := range lastMatchWord {
		message[i] = '*'
	}
	return message
}
