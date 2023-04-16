package anya

import (
	"testing"
)

func TestServer(t *testing.T) {
	s := NewHttpServer(":8085")
	err := s.Start()
	if err != nil {
		t.Fatal(err)
	}
}
