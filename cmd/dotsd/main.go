package main

import (
	"fmt"

	"github.com/innermond/dots/http"
)

const ADDR = ":8080"

func main() {
	fmt.Println("trying to start the server")
	server := http.NewServer()
	err := server.ListenAndServe(ADDR)
	fmt.Println(err)
}
