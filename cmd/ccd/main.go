package main

import (
	"log"

	"github.com/wayne930242/cc-dispatch/internal/daemon"
)

func main() {
	s, err := daemon.NewFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Serve(); err != nil {
		log.Fatal(err)
	}
}
