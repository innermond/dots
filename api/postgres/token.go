package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/innermond/dots"
	"github.com/innermond/dots/http/token"
)

type TokenService struct {
  db *DB
  tk token.Tokener

  ttl time.Duration
  prefix string

  userService *UserService
}

var (
  tokener token.Tokener
  errPrefix error = errors.New("token prefix not found")
)

func NewTokenService(db *DB, secret string, prefix string, ttl time.Duration, userService *UserService) *TokenService {
  if tokener == nil {
    tokener = token.Maker([]byte(secret))
  }

  return &TokenService{
    db: db, 
    tk: tokener,

    ttl: ttl,
    prefix: prefix,

    userService: userService,
  }
}

type loginData = dots.TokenCredentials

func (s *TokenService) Create(ctx context.Context, login loginData) (string, error) {
  if err := validateCreateFrom(login ); err != nil {
    return "", err
  }

  findByEmailApiKey := dots.UserFilter{
    Email: &login.Email,
    ApiKey: &login.Pass,
    Limit: 1,
  }
  uu, n, err := s.userService.FindUser(ctx, findByEmailApiKey)
  if err != nil {
    return "", err
  }
  if n == 0 {
    return "", errors.New("data not found")
  }
  uid := uu[0].ID

  tokenstr, err := s.tk.CreateToken(uid, s.ttl)
  if err != nil {
    return "", err
  }
  
  after, found := strings.CutPrefix(tokenstr, s.prefix)
  if !found {
    return "", errPrefix
  }
  
  return after, nil
}

func validateCreateFrom(data loginData) error {
  if data.Email == "" || data.Pass == "" {
    return errors.New("missing or invalid credentials")
  }
  return nil
}

func (s *TokenService) Read(ctx context.Context, str string) (*token.Payload, error) {
  str = s.prefix + str
  return s.tk.ReadToken(str)
}
