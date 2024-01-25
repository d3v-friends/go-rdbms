package mdPoint

import (
	"context"
	"fmt"
	"github.com/d3v-friends/go-rdbms/fnGorm"
	"github.com/d3v-friends/go-snippet/fn/fnEnv"
	"github.com/d3v-friends/go-snippet/fn/fnPanic"
	"github.com/d3v-friends/go-snippet/fn/fnParam"
	"gorm.io/gorm"
	"sync"
	"testing"
)

func TestAll(test *testing.T) {
	NewTestTool(true)
	TestAccount(test)
	TestCoupon(test)
	TestCouponUseRequest(test)
	TestWallet(test)
	TestWalletUseRequest(test)

}

type TestTool struct {
	DB *gorm.DB
}

func NewTestTool(truncate ...bool) (res *TestTool) {
	fnPanic.On(fnEnv.Load("../.env"))
	var db = fnPanic.Get(fnGorm.NewConnect(&fnGorm.ConnectArgs{
		Host:     fnEnv.GetString("DB_HOST"),
		Username: fnEnv.GetString("DB_USERNAME"),
		Password: fnEnv.GetString("DB_PASSWORD"),
		Schema:   fnEnv.GetString("DB_SCHEMA"),
	}))

	res = &TestTool{
		DB: db,
	}

	var models = []fnGorm.MigrateModel{
		&Account{},
		&Coupon{},
		&CouponBalance{},
		&CouponUseRequest{},
		&CouponUseReceipt{},
		&Wallet{},
		&WalletBalance{},
		&WalletUseRequest{},
		&WalletUseReceipt{},
	}

	if fnParam.Get(truncate) {
		res.TruncateAll(models)
	}

	fnPanic.On(fnGorm.RunMigrate(
		res.DB,
		models...,
	))

	return
}

func (x *TestTool) TruncateAll(models []fnGorm.MigrateModel) {
	x.DB.Exec("set FOREIGN_KEY_CHECKS = 0")
	for _, model := range models {
		x.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s", fnGorm.GetTableNm(model)))
	}
	x.DB.Exec("set FOREIGN_KEY_CHECKS = 1")
}

func (x *TestTool) Context() context.Context {
	var ctx = context.TODO()
	return fnGorm.SetDB(ctx, x.DB.WithContext(ctx))
}

func (x *TestTool) NewAccount() (res *Account) {
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	fnPanic.On(CreateAccount(x.Context(), &CreateAccountArgs{
		Fn: func(ctx context.Context, i *Account) (err error) {
			res = i
			wg.Done()
			return
		},
	}))
	wg.Wait()
	return
}
