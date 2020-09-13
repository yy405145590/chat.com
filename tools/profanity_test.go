package tools

import (
	"testing"
)

func init() {
	loadProfanity("../build/word.txt")
}
func Test_profanity(t *testing.T) {
	if (string(FilterMessage([]byte("hellboy")))) != "****boy" {
		t.Fail()
	}
	if (string(FilterMessage([]byte("helsboy")))) != "helsboy" {
		t.Fail()
	}
}
