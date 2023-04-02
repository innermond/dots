package dots

import (
	"context"

	"github.com/innermond/dots/autz"
)

func CanWriteOwn(ctx context.Context, tid int) error {
	user := UserFromContext(ctx)
	if user.ID == 0 {
		return Errorf(EUNAUTHORIZED, "unauthorized user")
	}

	canWriteOwn := autz.PowersContains(user.Powers, autz.WriteOwn) && user.ID == tid
	if !canWriteOwn {
		return Errorf(EUNAUTHORIZED, "unauthorized operation")
	}

	return nil
}
