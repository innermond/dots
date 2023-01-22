package dots

import "context"

type Ping struct {
	ID int `json:"id"`
}

type PingService interface {
	ById(ctx context.Context) *Ping
}
