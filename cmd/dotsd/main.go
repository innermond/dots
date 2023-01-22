package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/innermond/dots/http"
	"github.com/innermond/dots/postgres"
)

const addr = ":8080"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()

	dsn := "postgresql://postgres:admin@127.0.0.1:5432/dots?sslmode=disable"
	db := postgres.NewDB(dsn)
	if err := db.Open(); err != nil {
		panic(fmt.Errorf("cannot open database: %w", err))
	}

	pingService := postgres.NewPingService(db)
	server := http.NewServer()
	server.PingService = pingService

	go func() {
		fmt.Println("starting server...")
		log.Fatal(server.ListenAndServe(addr))
	}()

	<-ctx.Done()
}
