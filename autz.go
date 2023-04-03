package dots

import (
	"context"
)

func CanWriteOwn(ctx context.Context, tid int) error {
	user := UserFromContext(ctx)
	if user.ID == 0 {
		return Errorf(EUNAUTHORIZED, "unauthorized user")
	}

	canWriteOwn := PowersContains(user.Powers, WriteOwn) && user.ID == tid
	if !canWriteOwn {
		return Errorf(EUNAUTHORIZED, "unauthorized operation")
	}

	return nil
}

func CanReadOwn(ctx context.Context) error {
	user := UserFromContext(ctx)
	if user.ID == 0 {
		return Errorf(EUNAUTHORIZED, "unauthorized user")
	}

	canReadOwn := PowersContains(user.Powers, ReadOwn)
	if !canReadOwn {
		return Errorf(EUNAUTHORIZED, "unauthorized operation")
	}

	return nil
}
