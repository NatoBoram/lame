package main_test

import (
	"testing"

	. "github.com/NatoBoram/lame"
	"github.com/Sadzeih/go-reddit/reddit"
)

func TestFormatOpReply_Nil(t *testing.T) {
	result := FormatOpReply(nil)

	expected := ""
	if result != expected {
		t.Errorf("FormatOpReply(nil) = %s; expected %s", result, expected)
	}
}

func TestFormatOpReply(t *testing.T) {
	opReply := &reddit.Comment{Body: "Hello, world!"}
	result := FormatOpReply(opReply)

	expected := "Hello, world!"
	if result != expected {
		t.Errorf("FormatOpReply(%v) = %s; expected %s", opReply, result, expected)
	}
}
