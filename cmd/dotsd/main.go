package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/innermond/dots/http"
	"github.com/innermond/dots/postgres"
	"github.com/joho/godotenv"
)

const addr = ":8080"

func main() {
	fmt.Println("initiating...")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	clientId := os.Getenv("DOTS_GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("DOTS_GITHUB_CLIENT_SECRET")
	if clientId == "" || clientSecret == "" {
		log.Fatal("client credentials are missing")
	}

	dsn := os.Getenv("DOTS_DSN")
	db := postgres.NewDB(dsn)
	if err := db.Open(); err != nil {
		panic(fmt.Errorf("cannot open database: %w", err))
	}

	server := http.NewServer()

	err = server.OpenSecureCookie()
	if err != nil {
		log.Fatal(err)
	}

	server.ClientID = clientId
	server.ClientSecret = clientSecret

	authService := postgres.NewAuthService(db)
	userService := postgres.NewUserService(db)

	server.UserService = userService
	server.AuthService = authService

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()

	go func() {
		fmt.Println("starting server...")
		log.Fatal(server.ListenAndServe(addr))
	}()

	<-ctx.Done()

	log.Println("closing server...")

	if err := server.Close(); err != nil {
		log.Printf("shutdown: %w\n", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("shutdown: %w\n", err)
	}
}
