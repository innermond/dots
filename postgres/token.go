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

  ttl uint
  prefix string
}

var (
  tokener token.Tokener
  errPrefix error = errors.New("token prefix not found")
)

func NewTokenService(db *DB, secret string, prefix string, ttl uint) *TokenService {
  if tokener == nil {
    tokener = token.Maker([]byte(secret))
  }

  return &TokenService{
    db: db, 
    tk: tokener,

    ttl: ttl,
    prefix: prefix,
  }
}

func (s *TokenService) Create() (string, error) {
  d := time.Duration(s.ttl)*time.Second
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
