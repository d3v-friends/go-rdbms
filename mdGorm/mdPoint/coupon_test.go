package mdPoint

import (
	"context"
	"github.com/d3v-friends/go-snippet/fn/fnPanic"
	"github.com/d3v-friends/go-snippet/fn/fnReflect"
	"github.com/d3v-friends/go-snippet/typ"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCoupon(test *testing.T) {
	var tools = NewTestTool()
	var ctx = tools.Context()

	test.Run("createCoupon", func(t *testing.T) {
		var account = tools.NewAccount()

		// 쿠폰 생성 정보 체크
		fnPanic.On(CreateCoupon(ctx, &CreateCouponArgs{
			AccountId: account.Id,
			Currency:  decimal.NewFromInt(10000),
			Price:     decimal.NewFromInt(10),
			Fn: func(ctx context.Context, i *Coupon) (err error) {

				assert.Equal(t, "10000", i.Currency.String())
				assert.Equal(t, "10", i.Price.String())
				assert.Equal(t, "1000", i.Point.String())

				assert.Equal(t, "1000", i.CouponBalance.Point.String())
				assert.Equal(t, "10000", i.CouponBalance.Currency.String())
				return
			},
		}))

		// 반올림 체크
		fnPanic.On(CreateCoupon(ctx, &CreateCouponArgs{
			AccountId: account.Id,
			Currency:  decimal.NewFromInt(10000),
			Price:     decimal.NewFromInt(11),
			Fn: func(ctx context.Context, i *Coupon) (err error) {

				assert.Equal(t, "10000", i.Currency.String())
				assert.Equal(t, "11", i.Price.String())
				assert.Equal(t, "910", i.Point.String())

				assert.Equal(t, "910", i.CouponBalance.Point.String())
				assert.Equal(t, "10010", i.CouponBalance.Currency.String())
				return
			},
		}))
	})

	test.Run("findCoupon", func(t *testing.T) {
		var account = tools.NewAccount()
		var coupon *Coupon
		// 쿠폰 생성 정보 체크
		fnPanic.On(CreateCoupon(ctx, &CreateCouponArgs{
			AccountId: account.Id,
			Currency:  decimal.NewFromInt(10000),
			Price:     decimal.NewFromInt(10),
			Fn: func(ctx context.Context, i *Coupon) (err error) {
				assert.Equal(t, "10000", i.Currency.String())
				assert.Equal(t, "10", i.Price.String())
				assert.Equal(t, "1000", i.Point.String())

				coupon = i
				return
			},
		}))

		var loadedCoupons = fnPanic.Get(FindAllCouponsWithCtx(ctx, &FindCouponArgs{
			AccountId: []typ.UUID{
				account.Id,
			},
			HasBalance: fnReflect.Pointer(true),
		}))

		assert.Equal(t, 1, len(loadedCoupons))
		assert.Equal(t, "10000", loadedCoupons.RestCurrency().String())
		assert.Equal(t, "1000", loadedCoupons.RestPoint().String())
		assert.Equal(t, coupon.Id, loadedCoupons[0].Id)
	})
}
