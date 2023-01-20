package main

import (
	"fmt"

	"github.com/innermond/dots/http"
)

const ADDR = ":8080"

func main() {
	fmt.Println("server is starting")
	server := http.NewServer()
	err := server.ListenAndServe(ADDR)
	fmt.Println(err)
}
