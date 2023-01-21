package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/innermond/dots/http"
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

	fmt.Println("starting server...")
	server := http.NewServer()
	err := server.ListenAndServe(addr)
	fmt.Println(err)

	<-ctx.Done()
}
