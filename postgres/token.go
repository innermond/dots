package postgres

import (
	"errors"
	"strings"
	"time"

	"github.com/innermond/dots/http/token"
	"github.com/segmentio/ksuid"
)

type TokenService struct {
  db *DB
  tk token.Tokener
  prefix string
}

var (
  tokener token.Tokener
  errPrefix error = errors.New("token prefix not found")
)

func NewTokenService(db *DB, secret string) *TokenService {
  if tokener == nil {
    tokener = token.Maker([]byte(secret))
  }

  return &TokenService{
    db: db, 
    tk: tokener,
    prefix: "v4.local.",
  }
}

func (s *TokenService) Create() (string, error) {
  d := 1*time.Minute // TODO got it from a config or something
  uid := ksuid.New() // TODO get it from db
  
  tokenstr, err := s.tk.CreateToken(uid, d)
  if err != nil {
    return "", err
  }
  
  after, found := strings.CutPrefix(tokenstr, s.prefix)
  if !found {
    return "", errPrefix
  }
  
  return after, nil
}
