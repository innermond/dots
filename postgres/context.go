package postgres

import "context"

type keyTx string

var keytx keyTx = keyTx("tx")

func contextWithTx(ctx context.Context, tx *Tx) context.Context {
	return context.WithValue(ctx, keytx, tx)
}

func contextGetTx(ctx context.Context) *Tx {
	txMaybe := ctx.Value(keytx)
	if tx, ok := txMaybe.(*Tx); ok {
		return tx
	}
	return nil
}
