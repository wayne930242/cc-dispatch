package main

import (
	"log"

	"github.com/wayne930242/cc-dispatch/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
