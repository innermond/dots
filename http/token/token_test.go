package token

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/segmentio/ksuid"
)


func TestToken_Create(t *testing.T) {
  t.Run("create", func(t *testing.T) {

    err := godotenv.Load("../../.env")
    if err != nil {
      t.Fatal(err)
    }

    secret := os.Getenv("DOTS_TOKEN_SECRET")
    tokener := Maker([]byte(secret))
    
    uid := ksuid.New()
    d := 1 * time.Minute
    str, err := tokener.CreateToken(uid, d)
    if err != nil {
      t.Fatal(err)
    }

    payload, err := tokener.ReadToken(str)
    if err != nil {
      t.Fatal(err)
    }
    if payload.UID != uid {
      t.Fatalf("mismatch: %v != %v", payload.UID, uid)
    }
  
  })
} 
