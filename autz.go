package dots

import (
	"context"
)

func CanDoAnything(ctx context.Context) error {
	user := UserFromContext(ctx)
	if user.ID == 0 {
		return Errorf(EUNAUTHORIZED, "unauthorized user")
	}

	canDoAny := PowersContains(user.Powers, DoAnything)
	if canDoAny {
		return nil
	}

	return Errorf(EUNAUTHORIZED, "unauthorized operation")
}

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

func CanCreateOwn(ctx context.Context) error {
	user := UserFromContext(ctx)
	if user.ID == 0 {
		return Errorf(EUNAUTHORIZED, "unauthorized user")
	}

	canCreateOwn := PowersContains(user.Powers, CreateOwn)
	if !canCreateOwn {
		return Errorf(EUNAUTHORIZED, "unauthorized operation")
	}

	return nil
}