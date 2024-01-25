package mdPoint

import (
	"context"
	"github.com/d3v-friends/go-snippet/fn/fnPanic"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"sync"
	"testing"
)

func TestAccount(test *testing.T) {
	var tools = NewTestTool()
	var ctx = tools.Context()

	test.Run("create account", func(t *testing.T) {
		var account *Account
		var wg = &sync.WaitGroup{}
		wg.Add(1)
		fnPanic.On(CreateAccount(ctx, &CreateAccountArgs{
			Fn: func(ctx context.Context, i *Account) (err error) {
				account = i
				wg.Done()
				return
			},
		}))

		wg.Wait()

		assert.NotEmpty(t, account)
		assert.Equal(t, 1, len(account.Wallets))

		var rows *gorm.DB

		var couponCount int64
		if rows = tools.DB.
			Model(&Coupon{}).
			Where("`coupons`.`account_id` = ?", account.Id).
			Count(&couponCount); rows.Error != nil {
			t.Fatal(rows.Error)
		}
		assert.Equal(t, couponCount, int64(len(account.Coupons)))

		var walletCount int64
		if rows = tools.DB.
			Model(&Wallet{}).
			Where("`wallets`.`account_id` = ?", account.Id).
			Count(&walletCount); rows.Error != nil {
			t.Fatal(rows.Error)
		}

		assert.Equal(t, walletCount, int64(len(account.Wallets)))
	})
}
