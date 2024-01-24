package fnGorm

import (
	"context"
	"gorm.io/gorm"
)

type FnTrx func(context.Context, *gorm.DB) error

func Trx(ctx context.Context, fn FnTrx) error {
	var db = GetDBP(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		return fn(ctx, tx)
	})
}
