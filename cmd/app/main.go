package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	server := server.NewServer(ctx)
	done := make(chan bool, 1)

	go gracefulShutdown(server, done, cancel)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("err: server listen and serve: %v", err)
	}

	<-done
	log.Println("Graceful shutdown complete.")
}

func gracefulShutdown(apiServer *http.Server, done chan bool, fooCancel context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	fooCancel()

	fmt.Println()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")
	done <- true
}
