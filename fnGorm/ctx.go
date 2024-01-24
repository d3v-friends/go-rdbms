package fnGorm

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

const CtxGorm = "CTX_GORM_DB"

var (
	ErrNotFoundGormInContext = errors.New("not found gorm in context")
)

func GetDB(ctx context.Context) (db *gorm.DB, err error) {
	var isOk bool
	if db, isOk = ctx.Value(CtxGorm).(*gorm.DB); !isOk {
		err = ErrNotFoundGormInContext
		return
	}
	return
}

func GetDBP(ctx context.Context) (db *gorm.DB) {
	var err error
	if db, err = GetDB(ctx); err != nil {
		panic(err)
	}
	return
}

func SetDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, CtxGorm, db)
}
