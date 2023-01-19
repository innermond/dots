package main

import "net/http"

const ADDR = "localhost:8080"

func main() {
	http.ListenAndServe(ADDR)
}
