package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Print("server has started")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello\n")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
