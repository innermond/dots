package dots

import "context"

type TenentService interface {
	Set(context.Context) error
}
