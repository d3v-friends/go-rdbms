package mdPoint

import (
	"context"
	"fmt"
	"github.com/d3v-friends/go-snippet/fn/fnPanic"
	"github.com/d3v-friends/go-snippet/typ"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
	"testing"
)

func TestCouponUseRequest(test *testing.T) {
	var tools = NewTestTool()

	test.Run("shortage", func(t *testing.T) {
		var account = tools.NewAccount()
		var err = UseCoupon(tools.Context(), &UseCouponArgs{
			AccountId: account.Id,
			Point:     decimal.NewFromInt(10),
			Fn: func(ctx context.Context, req *CouponUseRequest) (err error) {
				return
			},
		})
		assert.Equal(t, ErrShortagePoint, err)
	})

	test.Run("request", func(t *testing.T) {
		var account = tools.NewAccount()
		// try 는 2로 나눌수 있는 수만 써야 한다!!
		var try int64 = 40
		var chargeCurrency int64 = 1000
		var maxPoint int64 = 100
		var price int64 = 10

		// test
		var wg = &sync.WaitGroup{}
		wg.Add(int(try))
		for i := 0; i < int(try); i++ {
			go func(tt *TestTool, w *sync.WaitGroup, a *Account, c int64, p int64) {
				fnPanic.On(CreateCoupon(tt.Context(), &CreateCouponArgs{
					AccountId: a.Id,
					Currency:  decimal.NewFromInt(c),
					Price:     decimal.NewFromInt(p),
					Fn: func(ctx context.Context, i *Coupon) (err error) {
						return
					},
				}))
				w.Done()
			}(tools, wg, account, chargeCurrency, price)
		}

		wg.Wait()

		var coupons = fnPanic.Get(FindAllCouponsWithCtx(tools.Context(), &FindCouponArgs{
			AccountId: []typ.UUID{account.Id},
		}))

		assert.Equal(t, fmt.Sprintf("%d", try*chargeCurrency), coupons.RestCurrency().String())

		wg = &sync.WaitGroup{}
		var totalUsePoint = decimal.Zero
		var halfTry = try / 2
		wg.Add(int(halfTry))
		for i := 0; i < int(halfTry); i++ {
			var usePoint = decimal.NewFromInt(rand.Int63n(maxPoint * 2))
			totalUsePoint = totalUsePoint.Add(usePoint)

			go func(tt *TestTool, w *sync.WaitGroup, p decimal.Decimal, a *Account) {
				fnPanic.On(UseCoupon(tt.Context(), &UseCouponArgs{
					AccountId: a.Id,
					Point:     p,
					Fn: func(ctx context.Context, req *CouponUseRequest) (err error) {
						return
					},
				}))
				w.Done()
			}(tools, wg, usePoint, account)
		}
		wg.Wait()

		var loadedCoupons = fnPanic.Get(FindAllCouponsWithCtx(tools.Context(), &FindCouponArgs{
			AccountId: []typ.UUID{account.Id},
		}))

		var totalUseCurrency = totalUsePoint.Mul(decimal.NewFromInt(price))
		var restCurrency = decimal.NewFromInt(try * chargeCurrency).Sub(totalUseCurrency)

		assert.Equal(t, restCurrency.String(), loadedCoupons.RestCurrency().String())
	})

	test.Run("cancel", func(t *testing.T) {
		var account = tools.NewAccount()
		var err = CreateCoupon(tools.Context(), &CreateCouponArgs{
			AccountId: account.Id,
			Currency:  decimal.NewFromInt(1000),
			Price:     decimal.NewFromInt(10),
			Fn: func(ctx context.Context, i *Coupon) (err error) {
				return fmt.Errorf("error")
			},
		})

		assert.Error(t, err)
	})
}
