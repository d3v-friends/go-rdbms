package fnGorm

import (
	"context"
	"gorm.io/gorm"
)

type FnTrxValue[T any] func(tx *gorm.DB) (res T, err error)

func Transaction[T any](ctx context.Context, fn FnTrxValue[T]) (res T, err error) {
	var db = GetDBP(ctx)
	var tx = db.
		Session(&gorm.Session{
			NewDB: true,
		}).
		WithContext(ctx).
		Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()

	res, err = fn(tx)
	return
}
