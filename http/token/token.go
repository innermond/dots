package token

import (
	"errors"
	"time"

	paseto "aidanwoods.dev/go-paseto"
	"github.com/innermond/dots"
	"github.com/segmentio/ksuid"
)

type Tokener interface{
  CreateToken(ksuid.KSUID, time.Duration) (string, error)
  ReadToken(string) (*Payload, error)
}

type Payload = dots.TokenPayload

func newPayload(uid ksuid.KSUID) *Payload {
  payload := Payload {
    ID: ksuid.New(),
    UID: uid,
  }

  return &payload
}

func Maker(key []byte) Tokener {
  return &pasetoMaker{
    key: key,
  }
}

type pasetoMaker struct {
  key []byte
}

func (k pasetoMaker) CreateToken(uid ksuid.KSUID, d time.Duration) (string, error) {
  token := paseto.NewToken()
  payload := newPayload(uid)
  now := time.Now().UTC()
  token.SetIssuedAt(now)
  token.SetNotBefore(now)
  exp := time.Now().Add(d)
  token.SetExpiration(exp)
  token.Set("payload", payload)

  sk, err := paseto.V4SymmetricKeyFromBytes(k.key)
  if err != nil {
    return "", err
  }

  return token.V4Encrypt(sk, nil), nil
}

type ErrTokenClient = paseto.RuleError

func (k pasetoMaker) ReadToken(token string) (*Payload, error) {
  sk, err := paseto.V4SymmetricKeyFromBytes(k.key)
  if err != nil {
    return nil, err
  }

  p := paseto.NewParser()
  tok, err := p.ParseV4Local(sk, token, nil)
  if err != nil {
    if errors.Is(err, &ErrTokenClient{}) {
      perr := &dots.Error{Code: dots.EUNAUTHORIZED, Message: "unauthorized token",}
      return nil, perr.Wrap(err)
    }
    return nil, err
  }

  var payload Payload
  err = tok.Get("payload", &payload)
  if err != nil {
    return nil, err
  }

  return &payload, nil
}
