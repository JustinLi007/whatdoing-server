package main

import (
	"log"

	"github.com/JustinLi007/whatdoing-server/internal/server"
)

func main() {
	server := server.NewServer()
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("err: server listen and serve: %v", err)
	}
}
